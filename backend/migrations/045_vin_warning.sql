-- +goose Up
-- Match the entry form's standard U.S. VIN check. Blank VINs are optional.
-- +goose StatementBegin
CREATE FUNCTION vin_has_warning(value text) RETURNS boolean
LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE AS $$
DECLARE
    vin text := upper(regexp_replace(coalesce(value, ''), '[[:space:]]', '', 'g'));
    alphabet text := 'ABCDEFGHJKLMNPRSTUVWXYZ';
    letter_values integer[] := ARRAY[1,2,3,4,5,6,7,8,1,2,3,4,5,7,9,2,3,4,5,6,7,8,9];
    weights integer[] := ARRAY[8,7,6,5,4,3,2,10,0,9,8,7,6,5,4,3,2];
    total integer := 0;
    character text;
    numeric_value integer;
    remainder integer;
BEGIN
    IF vin = '' THEN RETURN false; END IF;
    IF length(vin) <> 17 OR vin !~ '^[A-HJ-NPR-Z0-9]+$' THEN RETURN true; END IF;
    FOR position IN 1..17 LOOP
        character := substr(vin, position, 1);
        IF character ~ '^[0-9]$' THEN
            numeric_value := character::integer;
        ELSE
            numeric_value := letter_values[strpos(alphabet, character)];
        END IF;
        total := total + numeric_value * weights[position];
    END LOOP;
    remainder := total % 11;
    RETURN substr(vin, 9, 1) <> CASE WHEN remainder = 10 THEN 'X' ELSE remainder::text END;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION vin_has_warning(text);

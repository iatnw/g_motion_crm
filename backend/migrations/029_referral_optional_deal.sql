-- +goose Up
ALTER TABLE referrals ALTER COLUMN deal_id DROP NOT NULL;

-- +goose Down
ALTER TABLE referrals ALTER COLUMN deal_id SET NOT NULL;

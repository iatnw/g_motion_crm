type CommissionFields = {
  broker_fee: string;
  collect_amount: string;
  doc_fee: string;
  broker_percent: string;
  broker_commission: string;
  your_net: string;
};

// Repair legacy records that only stored one commission field. Preserve records
// with both fields populated, including any intentional adjustments to the net.
export function fillMissingCommission<T extends CommissionFields>(form: T, referralTotal: number): T {
  const base = (parseFloat(form.broker_fee) || 0) + (parseFloat(form.collect_amount) || 0) - (parseFloat(form.doc_fee) || 0);
  const percent = parseFloat(form.broker_percent);
  const commission = parseFloat(form.broker_commission);
  if (base <= 0) return form;
  if (percent > 0 && !(commission > 0)) {
    const computed = base * percent / 100;
    return { ...form, broker_commission: computed.toFixed(2), your_net: (base - computed - referralTotal).toFixed(2) };
  }
  if (commission > 0 && !(percent > 0)) {
    return { ...form, broker_percent: (commission / base * 100).toFixed(2), your_net: (base - commission - referralTotal).toFixed(2) };
  }
  return form;
}

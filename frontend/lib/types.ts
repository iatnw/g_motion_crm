export type ClientStatus = "active" | "closed" | "no_deal";
export type ClientSource =
  | "instagram"
  | "telegram"
  | "referall"
  | "organic"
  | "repeat_client";
export type EmploymentStatus = "self_employed" | "full_time" | "part_time";
export type DocumentType =
  | "credit_app_pdf"
  | "driver_license"
  | "insurance"
  | "broker_agreement"
  | "dealer_docs"
  | "check_copy"
  | "other";
export type PaymentStatus = "paid" | "pending";
export type ReferralStatus = "none" | "paid" | "pending";
export type DownPaymentMethod = "cash" | "card";
export type DealStatus = "paid" | "pending" | "cancelled";
export type BrokerPaymentMethod = "zelle_gmotion" | "zelle_personal" | "cash";
export type PendingPaymentBucket = "green" | "yellow" | "red";

export interface Client {
  id: string;
  first_name: string;
  last_name: string;
  phone: string;
  email: string;
  status: ClientStatus;
  source?: ClientSource;
  notes?: string;
  date_of_birth?: string;
  address?: string;
  city?: string;
  state?: string;
  zip?: string;
  ssn?: string;
  housing_status?: string;
  housing_duration?: string;
  housing_monthly_payment?: string;
  employer_name?: string;
  employer_address?: string;
  employer_city?: string;
  employer_state?: string;
  employer_zip?: string;
  occupation?: string;
  employer_phone?: string;
  employment_status?: EmploymentStatus;
  employment_duration?: string;
  gross_monthly_income?: string;
  other_monthly_income?: string;
  other_income_source?: string;
  created_at: string;
  updated_at: string;
}

export interface Broker {
  id: string;
  first_name: string;
  last_name: string;
  phone?: string;
  email?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Dealer {
  id: string;
  name: string;
  contact_person?: string;
  contact_person_email?: string;
  phone?: string;
  address?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Deal {
  id: string;
  client_id: string;
  broker_id?: string;
  dealer_id?: string;
  deal_date?: string;
  deal_number?: string;
  vin?: string;
  make?: string;
  model?: string;
  trim_level?: string;
  down_payment?: number;
  down_payment_method?: DownPaymentMethod;
  monthly_payment?: number;
  broker_fee?: number;
  broker_percent?: number;
  broker_commission?: number;
  broker_commission_status?: PaymentStatus;
  broker_commission_paid_date?: string;
  your_net?: number;
  broker_fee_status?: PaymentStatus;
  broker_fee_paid_date?: string;
  deal_status?: DealStatus;
  deal_status_pending_since?: string;
  pending_payment_days?: number;
  pending_payment_bucket?: PendingPaymentBucket;
  payment_reminder_email_sent_at?: string;
  collect_amount?: number;
  doc_fee?: number;
  broker_payment_method?: BrokerPaymentMethod;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Referral {
  id: string;
  deal_id: string;
  first_name: string;
  last_name: string;
  phone?: string;
  amount?: number;
  status: ReferralStatus;
  paid_date?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface DealShow {
  id: string;
  date?: string;
  client_name: string;
  make?: string;
  model?: string;
  trim?: string;
  vin?: string;
  dealer_name?: string;
  broker_name?: string;
  broker_fee?: number;
  collect_amount?: number;
  doc_fee?: number;
  referral_total?: number;
  broker_percent?: number;
  broker_profit?: number;
  company_net?: number;
  check_status?: PaymentStatus;
  broker_paid_status?: PaymentStatus;
  broker_payment_type?: BrokerPaymentMethod;
  notes?: string;
}

export interface DashboardSummary {
  totals: {
    total_deals: number;
    gross_revenue?: number;
    broker_payout?: number;
    total_doc_fees?: number;
    company_profit_before_tax?: number;
    paid_revenue?: number;
    pending_revenue?: number;
    paid_broker_payout?: number;
    unpaid_broker_payout?: number;
    pending_company_profit?: number;
    avg_revenue_per_deal?: number;
    avg_profit_per_deal?: number;
    total_referral_payout?: number;
  };
  aging: {
    pending_0_31_count: number;
    pending_0_31_revenue?: number;
    pending_32_40_count: number;
    pending_32_40_revenue?: number;
    pending_41_plus_count: number;
    pending_41_plus_revenue?: number;
    oldest_pending_days: number;
  };
  brokers: Array<{
    id: string;
    first_name: string;
    last_name: string;
    deal_count: number;
    gross_revenue?: number;
    broker_payout?: number;
    company_profit_before_tax?: number;
    avg_revenue_per_deal?: number;
  }>;
  dealers: Array<{
    id: string;
    name: string;
    deal_count: number;
    gross_revenue?: number;
    avg_days_to_payment?: number;
  }>;
  recent_deals: Deal[];
}

export interface Document {
  id: string;
  deal_id?: string;
  client_id?: string;
  s3_key: string;
  file_name: string;
  file_size_bytes?: number;
  document_type: DocumentType;
  uploaded_at: string;
}

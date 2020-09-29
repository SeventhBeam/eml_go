package eml

import (
	"encoding/json"
	"fmt"
)

type ErrorModel struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

func (e *ErrorModel) Error() string {
	return fmt.Sprintf("EML %s: %s", e.Code, e.Description)
}

func (e *ErrorModel) UserMessage() string {
	return e.Description
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type CreateAccountRequest struct {
	UpdateAccountRequest
	CompanyId              json.Number `json:"company_id"`
	InitialLoadAmount      json.Number `json:"initial_load_amount"`
	NameOnCard             string      `json:"name_on_card,omitempty"`
	IsPlasticEnabled       bool        `json:"is_plastic_enabled"`
	PlasticExpiry          string      `json:"plastic_expiry,omitempty"`
	CorrespondingAccountId string      `json:"corresponding_account_id,omitempty"`
	CardHolderType         int         `json:"card_holder_type,omitempty"`
}

type UpdateAccountRequest struct {
	ClientAccountKey string            `json:"client_account_key,omitempty"`
	InitiatingUserId string            `json:"initiating_user_id,omitempty"`
	Registration     *Registration     `json:"registration"`
	PortalIdentifier *PortalIdentifier `json:"portal_identifier,omitempty"`
	MdesConfigId     string            `json:"mdes_config_id,omitempty"`
	AccountExpiry    string            `json:"account_expiry,omitempty"`
}

type PortalIdentifier struct {
	ClientId     string `json:"client_id"`
	ProgramId    string `json:"program_id"`
	CardholderId string `json:"cardholder_id"`
}

type Registration struct {
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	PrimaryAddress   *Address `json:"primary_address"`
	AlternateAddress *Address `json:"alternate_address,omitempty"`
	MobileNumber     string   `json:"mobile_number,omitempty"`
	DateOfBirth      string   `json:"date_of_birth"`
	EmailAddress     string   `json:"email_address"`
}

type Address struct {
	Line1    string `json:"address_line1"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`
}

type AccountSummary struct {
	Balance           json.Number `json:"balance"`
	CardNumber        string      `json:"card_number"`
	CompanyId         json.Number `json:"company_id"`
	ExternalAccountId string      `json:"external_account_id"`
	PlasticExpiry     string      `json:"plastic_expiry"`
	ProductType       string      `json:"product_type"`
	Status            CardStatus  `json:"status"`
	IsPlasticEnabled  bool        `json:"is_plastic_enabled"`
}

func (a *AccountSummary) GetName() string {
	return "EML Account"
}

func (a *AccountSummary) GetStatus() string {
	return string(a.Status)
}

type AccountInfo struct {
	*AccountSummary
	AccountId                string            `json:"account_id"`
	AccountExpiry            string            `json:"account_expiry"`
	BpayBillerCode           string            `json:"bpay_biller_code"`
	BpayRefNumber            string            `json:"bpay_reference_number"`
	DirectEntryBsb           string            `json:"direct_entry_bsb"`
	DirectEntryAccountNumber string            `json:"direct_entry_account_number"`
	Personal                 *Registration     `json:"personal"`
	PortalIdentifier         *PortalIdentifier `json:"portal_identifier"`
	MdesConfigId             string            `json:"mdes_config_id"`
}

type TransactionsPage struct {
	NextCursor string
	PageSize   int
	Count      int
	Items      []Transaction
}

type Transaction struct {
	Id         string      `json:"id"`
	ParentId   string      `json:"parent_id"`
	OccurredAt string      `json:"occurred_at"`
	Amount     json.Number `json:"amount"`
	Reference  string      `json:"reference"`
}

type StatusRequest struct {
	Status CardStatus `json:"status"`
}

type PlasticEnabledRequest struct {
	PlasticEnabled bool `json:"plastic_enabled"`
}

type TransferRequest struct {
	Amount               json.Number `json:"amount"`
	SourceReference      string      `json:"source_reference"`
	DestinationReference string      `json:"destination_reference"`
	DestinationAccountId string      `json:"destination_account_id"`
	TransactionType      int         `json:"transaction_type"`
	Username             string      `json:"initiator_username"`
	RequestId            string      `json:"request_id"`
}

type RegistrationInfo interface {
	Uid() string
	Email() string
	Phone() string
	FirstName() string
	LastName() string
}

type freeFieldsRequest struct {
	Text1 string  `json:"free_text1"`
	Text2 string  `json:"free_text2"`
	Text3 string  `json:"free_text3"`
	Text4 string  `json:"free_text4"`
	Text5 string  `json:"free_text5"`
	Text6 string  `json:"free_text6"`
	Text7 string  `json:"free_text7"`
	Text8 string  `json:"free_text8"`
	Int1  int     `json:"free_int1"`
	Int2  int     `json:"free_int2"`
	Dec1  float64 `json:"free_dec1"`
	Dec2  float64 `json:"free_dec2"`
}

type HookPage struct {
	NextCursor string
	PageSize   int
	Count      int
	Items      []Hook
}

type Hook struct {
	Id                         string `json:"id"`
	Uri                        string `json:"uri"`
	Scope                      []int  `json:"scope"`
	FilterSpec                 string `json:"filter_spec"`
	Enabled                    bool   `json:"enabled"`
	ReliabilityMode            string `json:"reliability_mode"`
	LastUndeliverable          string `json:"last_undeliverable"`
	LastUndeliverableTimestamp string `json:"last_undeliverable_timestamp"`
	HmacKeyId                  string `json:"hmac_key_id"`
}

type HookRequest struct {
	Uri             string          `json:"uri,omitempty"`
	Scope           []int           `json:"scope,omitempty"`
	FilterSpec      string          `json:"filter_spec,omitempty"`
	Enabled         *bool           `json:"enabled,omitempty"`
	ReliabilityMode ReliabilityMode `json:"reliability_mode,omitempty"`
	HmacKeyId       string          `json:"hmac_key_id,omitempty"`
	/* You must provide a 256bit hex value. (i.e., exactly 64 hex characters). */
	HmacKeySecret string `json:"hmac_key_secret,omitempty"`
}

type IdModel struct {
	Id string `json:"id"`
}

type MessageIdsRequest struct {
	MessageIds []string `json:"message_ids"`
}

type MessagePage struct {
	More     bool
	PageSize int
	Items    []Message
}

type Message struct {
	Id                string      `json:"id"`
	HookId            string      `json:"hook_id"`
	HookManagementUri string      `json:"hook_management_uri"`
	Timestamp         string      `json:"timestamp"`
	Type              string      `json:"type"`
	Version           string      `json:"version"`
	Data              MessageData `json:"data"`
}

type MessageData struct {
	// UndeliverableAlert
	LastUndeliverable          string `json:"last_undeliverable"`
	LastUndeliverableTimestamp string `json:"last_undeliverable_timestamp"`

	/* The id of the logical transaction which is new or updated in this message. */
	LogicalTransactionId int64 `json:"logical_transaction_id"`
	/* The External Account Id of the account on which this transaction has occurred. */
	AccountId    string          `json:"account_id"`
	CustomerInfo TxnCustomerInfo `json:"customer_info"`
	CompanyId    int             `json:"company_id"`
	/* Where is this logical transaction in its transaction lifecycle? */
	StateChange TxnStateChange `json:"state_change"`
	Details     TxnDetails     `json:"details"`
	/* More information about this specific event. */
	Event TxnEvent `json:"event"`
}

type TxnCustomerInfo struct {
	ClientId         int    `json:"client_id"`
	ClientAccountKey string `json:"client_account_key"`
}

type TxnStateChange struct {
	Lifecycle string   `json:"lifecycle"`
	OldState  TxnState `json:"old_state"`
	NewState  TxnState `json:"new_state"`
	/* Is this the first message containing this logical_transaction_id? */
	IsNew bool `json:"is_new"`
	/* Does this state change indicate an adjustment to the previous value of the transaction? */
	IsAdjustment bool `json:"is_adjustment"`
}

type TxnDetails struct {
	/* The date/time of the original transaction */
	OccurredAt string `json:"occurred_at"`
	/* Where did the transaction originate. */
	Source       string `json:"source"`
	MobileWallet string `json:"mobile_wallet"`
	/*
	 * The base amount of the transaction, as charged by the merchant
	 * and displayed on the receipt.
	 *
	 * Includes any cash_amount for a ATM or Cash-out-at-POS transaction.
	 * Does not include internal or external fee amounts.
	 */
	BaseAmount TxnCurrency `json:"base_amount"`
	/*
	 * The cash component of an ATM Withdrawal or a Cash-out-at-POS transaction.
	 *
	 * Always less than or equal to the base_amount.
	 */
	CashAmount TxnCurrency `json:"cash_amount"`
	/*
	 * Fee amount charged by EML to the account holder on this transaction.
	 *
	 * If there are fee(s) on this transaction which attract GST,
	 * the amount shown here is inclusive of that tax.
	 */
	InternalFeeAmount TxnCurrency `json:"internal_fee_amount"`
	/*
	 * External fees charged to the account holder on this transaction.
	 * e.g., ATM owner fees
	 */
	ExternalFeeAmount TxnCurrency `json:"external_fee_amount"`
	/*
	 * (Optional) Details about the merchant/acquirer currency.
	 * This sub-element will be null for domestic transactions.
	 */
	ForeignExchange *TxnForex `json:"foreign_exchange"`
	/* Merchant/transaction reference. */
	Description string `json:"description"`
	/*
	 * (Optional) Merchant particulars.
	 * This sub-element may be null for some transactions.
	 */
	Merchant *TxnMerchant `json:"merchant"`
	/*
	 * (Optional) Alternate transaction identifiers.
	 * This sub-element may be null for some transactions.
	 */
	Identifiers *TxnIdentifiers `json:"identifiers"`
	/* Is the transaction is currently reversed, cancelled or declined? */
	IsVoid bool `json:"is_void"`
	/*
	 * Does this transaction indicate a changed account status?
	 * e.g. An activation/deactivation/etc.
	 */
	IsAccountStatusChange bool `json:"is_account_status_change"`
	/*
	 * (Optional) The reason why the transaction was declined.
	 *
	 * One of:  unknown = 0
	 *			insufficient_funds = 1
	 *			incorrect_pin = 2
	 *			merchant_disallowed = 3
	 *          velocity_exceeded = 4
	 *			account_inactive = 5
	 *			account_expired = 6
	 *			system_error = 7
	 *			other = 8
	 *
	 * If it is not a declined transaction, the value is null.
	 */
	DeclineReason TxnDeclineReason `json:"decline_reason"`
}

type TxnCurrency struct {
	/*
	 * The amount, in the minor currency unit.
	 * i.e., in cents/pence/etc. (-3750 = -$37.50)
	 */
	ValueMinor int `json:"value_minor"`
	/*
	 * An ISO 4217 numeric currency code.
	 * e.g., 036 = Australian Dollar, 840 = US Dollar.
	 */
	Currency string `json:"currency"`
}

type TxnForex struct {
	/*
	 * The source/merchant currency.
	 * An ISO 4217 numeric currency code.
	 */
	SourceCurrency string `json:"source_currency"`
	/*
	 * The conversion rate between the source currency and the account
	 * holder's currency.
	 *
	 * base_amount = merchant_price * conversion_rate
	 */
	ConversionRate json.Number `json:"conversion_rate"`
}

type TxnMerchant struct {
	/* (Optional) Category Code */
	MerchantCategory string `json:"merchant_category"`
	/* (Optional) Acquiring institution identification code */
	AcquirerId int `json:"acquirer_id"`
	/* (Optional) Card Acceptor Id (a.k.a. Merchant Id) */
	CardAcceptorId string `json:"card_acceptor_id"`
	/* (Optional) Terminal Id */
	TerminalId string `json:"terminal_id"`
	/* (Optional) Network supplied merchant name and location */
	CardAcceptorNameLocation string `json:"card_acceptor_name_location"`
}

type TxnIdentifiers struct {
	/* (Optional) Authorisation Id */
	AuthId string `json:"auth_id"`
	/* (Optional) System Trace Audit Number */
	Stan string `json:"stan"`
	/* (Optional) Retrieval Reference Number */
	Rrn string `json:"rrn"`
	/*
	 * (Optional) Network specific transaction clearing identifier.
	 * e.g., Visa Transaction Id
	 */
	ClearingId string `json:"clearing_id"`
}

type TxnEvent struct {
	/* The sequence number of the corresponding transaction history item. */
	Id int64 `json:"id"`
	/* The EML transaction type code. */
	TypeCode string `json:"type_code"`
	/* The date and time of the event. */
	Timestamp string `json:"timestamp"`
	/*
	 * (Optional) The Id of the delegated transaction decision.
	 *
	 * If this event represents a transaction decision which was delegated
	 * to a third party, this is the id of that request.
	 */
	DelegationRequestId string `json:"delegation_request_id"`
	/* The net change in balance on the account due to this event. */
	BalanceDelta TxnCurrency `json:"balance_delta"`
	/* The balance on the account after the event. */
	RunningBalance TxnCurrency `json:"running_balance"`
}

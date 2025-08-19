package eml

type GetAccountFlag string

const (
	pathToken          = "/3.0/token"
	contentTypeEmlJson = "application/vnd.eml+json"
	contentTypeJson    = "application/json"
	headerAccept       = "Accept"
	headerContentType  = "Content-Type"
)

const (
	WithPersonal    GetAccountFlag = "with_personal"
	WithDirectEntry GetAccountFlag = "with_directentry"
	WithFreeText    GetAccountFlag = "with_freetext"
	WithBpay        GetAccountFlag = "with_bpay"
	WithTokenInfo   GetAccountFlag = "with_tokeninfo"
)
const (
	TransactionTypeCardToCard = 2902
)
const (
	AddressTypeIndividual = "individual"
	AddressTypeBatch      = "batch"
)
const (
	headerPageSize   = "X-PageSize"
	headerTotalPages = "X-TotalPages"
	headerTotalItems = "X-Totalitems"
)
const (
	queryPageSize       = "page_size"
	queryFromDate       = "start_date"
	queryToDate         = "end_date"
	queryPageNumber     = "page_number"
	transactionViewType = "view_type"
)

type TransactionViewType string

const (
	DefaultView    TransactionViewType = "default"
	SimplifiedView TransactionViewType = "simplified"
)

type TransactionType string

const (
	TxnTypeTransaction        = "transaction"
	TxnTypePing               = "ping"
	TxnTypeUndeliverableAlert = "undeliverable_alert"
)
const (
	LifecycleSimple         = "simple"
	LifecyclePaymentNetwork = "payment_network"
)
const (
	SourceMastercard = "mastercard"
	SourceVisa       = "visa"
	SourceEftpos     = "eftpos"
	SourceBlackhawk  = "blackhawk"
	SourceEpay       = "epay"
	SourceIncomm     = "incomm"
)
const (
	WalletApple   = "apple"
	WalletGoogle  = "google"
	WalletSamsung = "samsung"
	WalletNone    = ""
)
const (
	CurrencyAud = "036"
	CurrencyUsd = "840"
)

type TxnDeclineReason string

const (
	DeclineUnknown            TxnDeclineReason = "unknown"
	DeclineInsufficientFunds  TxnDeclineReason = "insufficient_funds"
	DeclineIncorrectPin       TxnDeclineReason = "incorrect_pin"
	DeclineMerchantDisallowed TxnDeclineReason = "merchant_disallowed"
	DeclineVelocityExceeded   TxnDeclineReason = "velocity_exceeded"
	DeclineAccountInactive    TxnDeclineReason = "account_inactive"
	DeclineAccountExpired     TxnDeclineReason = "account_expired"
	DeclineSystemError        TxnDeclineReason = "system_error"
	DeclineOther              TxnDeclineReason = "other"
)

type ReliabilityMode string

const (
	StoreUndeliverable ReliabilityMode = "store_undeliverable"
	None               ReliabilityMode = "none"
)
const FilterSpecAll = "*"

const (
	UpdatePlasticEnabled = "3002"
	UpdateCardStatus     = "3003"
)

type CardStatus string

const (
	CardStatusPending                  CardStatus = "pending"
	CardStatusActive                   CardStatus = "active"
	CardStatusPreActive                CardStatus = "pre_active"
	CardStatusInactive                 CardStatus = "inactive"
	CardStatusDeactivated              CardStatus = "deactivated"
	CardStatusLostStolen               CardStatus = "lost_or_stolen"
	CardStatusExpired                  CardStatus = "expired"
	CardStatusFraud                    CardStatus = "suspected_fraud"
	CardStatusClosed                   CardStatus = "closed"
	CardStatusInactivePinTriesExceeded CardStatus = "inactive_pin_tries_exceeded"
	CardStatusEmlInactive              CardStatus = "eml_inactive"
	CardStatusReplaced                 CardStatus = "replaced"
)

type TxnState string

const (
	StateNew        TxnState = "new"
	StateRequested  TxnState = "requested"
	StateAuthorized TxnState = "authorized"
	StateCleared    TxnState = "cleared"
	StatePosted     TxnState = "posted"
	StateDeclined   TxnState = "declined"
	StateCancelled  TxnState = "cancelled"
	StateReversed   TxnState = "reversed"
)

type CommunicateMethod string

const (
	CommunicateMethodSMS   CommunicateMethod = "sms"
	CommunicateMethodEmail CommunicateMethod = "email"
)

type OperationType string

const (
	OperationTypeOneTimePassCode OperationType = "one_time_pass_code"
)

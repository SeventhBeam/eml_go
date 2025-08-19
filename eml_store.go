package eml

import (
	"context"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Store interface {
	CreateAccount(ctx context.Context, request *CreateAccountRequest) (*AccountSummary, error)
	GetAccount(ctx context.Context, eaid string, flags ...GetAccountFlag) (*AccountInfo, error)
	GetSummary(ctx context.Context, eaid string) (*AccountSummary, error)
	GetTransactions(ctx context.Context, eaid string, pageSize int, cursor string, startDate *time.Time, endDate *time.Time) (*TransactionsPage, error)
	UpdateStatus(ctx context.Context, eaid string, status CardStatus) error
	UpdatePlasticEnabled(ctx context.Context, eaid string, enabled bool) error
	UpdateRegistration(ctx context.Context, eaid string, info RegistrationInfo) error
	//UpdateFreeFields(ctx context.Context, eaid string, fields *FreeFields) error
	Transfer(ctx context.Context, eaid string, request *TransferRequest) error

	// Notifications
	AddHook(ctx context.Context, model *HookRequest) (string, error)
	GetHooks(ctx context.Context) (*HookPage, error)
	GetHook(ctx context.Context, hookId string) (*Hook, error)
	DeleteHook(ctx context.Context, hookId string) error
	UpdateHookScope(ctx context.Context, hookId string, scope []int) error
	GetUndeliverable(ctx context.Context, hookId string, pageSize int, pageNumber int) (*MessagePage, error)
	DismissUndeliverable(ctx context.Context, hookId string, messageIds []string) error

	// Renewals
	Authenticate(ctx context.Context, eaid string, req AuthenticateRequest) (*AuthenticateResponse, error)
	Initiate(ctx context.Context, eaid string, req InitiateRequest) (*InitiateResponse, error)
	Activate(ctx context.Context, eaid string, req ActivateRequest) error
}

type Settings struct {
	FunctionHost string
	HookUri      string
	EmlRestId    string
	EmlHostUrl   string
	DebugRest    bool
}

type emlStore struct {
	_restSecret string
	_env        *Settings
	_lazyClient *resty.Client
	_lazyOnce   sync.Once
	token       *bearerToken
	refreshing  uint32
}

func (e *emlStore) lazyInit(ctx context.Context) {
	e._lazyOnce.Do(func() {
		e._lazyClient = resty.New().
			SetHostURL(e._env.EmlHostUrl).
			SetBasicAuth(e._env.EmlRestId, e._restSecret).
			SetHeader(headerAccept, contentTypeEmlJson).
			SetError(&ErrorModel{}).
			SetDebug(e._env.DebugRest).
			OnBeforeRequest(e.onBeforeRequest).
			OnBeforeRequest(logRequest).
			OnAfterResponse(logResponse)
	})
}

func (e *emlStore) request(ctx context.Context) *resty.Request {
	e.lazyInit(ctx)
	return e._lazyClient.R().SetContext(ctx)
}

func (e *emlStore) CreateAccount(ctx context.Context, account *CreateAccountRequest) (*AccountSummary, error) {
	log.Printf("Creating EML account in company %s with load %s from %s", account.CompanyId, account.InitialLoadAmount, account.CorrespondingAccountId)
	resp, err := e.request(ctx).
		SetBody(account).
		SetResult(&AccountSummary{}).
		SetHeader(headerContentType, contentTypeEmlJson).
		Post("/3.0/accounts")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*AccountSummary), nil
}

func (e *emlStore) GetAccount(ctx context.Context, eaid string, flags ...GetAccountFlag) (*AccountInfo, error) {
	log.Printf("Getting account %s with flags %v", eaid, flags)
	queryParams := make(map[string]string)
	for _, flag := range flags {
		queryParams[string(flag)] = "1"
	}
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetQueryParams(queryParams).
		SetResult(&AccountInfo{}).
		Get("/3.0/accounts/{id}")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*AccountInfo), nil
}

func (e *emlStore) GetSummary(ctx context.Context, eaid string) (*AccountSummary, error) {
	log.Printf("Getting account %s summary", eaid)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetResult(&AccountSummary{}).
		Get("/3.0/accounts/{id}/status")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*AccountSummary), nil
}

func (e *emlStore) GetTransactions(ctx context.Context, eaid string, pageSize int, cursor string, startDate *time.Time, endDate *time.Time) (*TransactionsPage, error) {
	log.Printf("Getting account %s transactions (%d, %s)", eaid, pageSize, cursor)
	pageNumber := cursor
	if pageNumber == "" {
		pageNumber = "1"
	}
	//2018-02-24T09:02:10Z
	df := "2006-01-02T03:04:05Z"
	startDateStr := ""
	if startDate != nil {
		startDateStr = startDate.Format(df)
	}
	endDateStr := ""
	if endDate != nil {
		endDateStr = endDate.Format(df)
	}

	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetQueryParams(map[string]string{
			queryPageNumber: pageNumber,
			queryPageSize:   strconv.Itoa(pageSize),
			queryFromDate:   startDateStr,
			queryToDate:     endDateStr,
		}).
		SetResult([]Transaction{}).
		Get("/3.0/accounts/{id}/transactions")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	actualPageSize, _ := strconv.Atoi(resp.Header().Get(headerPageSize))
	totalPages, _ := strconv.Atoi(resp.Header().Get(headerTotalPages))
	totalItems, _ := strconv.Atoi(resp.Header().Get(headerTotalItems))
	nextPageNumber, _ := strconv.Atoi(pageNumber)
	nextPageNumber++
	nextCursor := ""
	if nextPageNumber <= totalPages {
		nextCursor = strconv.Itoa(nextPageNumber)
	}
	var txns []Transaction
	if resp.StatusCode() == http.StatusNoContent {
		txns = []Transaction{}
	} else {
		txns = *resp.Result().(*[]Transaction)
	}
	page := &TransactionsPage{
		//TotalPages: totalPages,
		TotalItems: totalItems,
		PageSize:   actualPageSize,
		NextCursor: nextCursor,
		Count:      len(txns),
		Items:      txns,
	}
	return page, nil
}

func (e *emlStore) UpdateStatus(ctx context.Context, eaid string, status CardStatus) error {
	log.Printf("Updating account %s status %s", eaid, status)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(StatusRequest{Status: status}).
		SetHeader(headerContentType, contentTypeEmlJson).
		Put("/3.0/accounts/{id}/status")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) UpdatePlasticEnabled(ctx context.Context, eaid string, enabled bool) error {
	log.Printf("Updating account %s isPlasticEnabled %v", eaid, enabled)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(PlasticEnabledRequest{PlasticEnabled: enabled}).
		SetHeader(headerContentType, contentTypeEmlJson).
		Put("/3.0/accounts/{id}/plastic")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) UpdateRegistration(ctx context.Context, eaid string, info RegistrationInfo) error {
	log.Printf("Updating account %s registration %s %s\n", eaid, info.Email(), info.Phone())
	account, err := e.GetAccount(ctx, eaid, WithPersonal)
	if err != nil {
		return ContextualError(err, "e.GetAccount")
	}
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(mapRegistrationUpdate(account, info)).
		SetHeader(headerContentType, contentTypeEmlJson).
		Put("/3.0/accounts/{id}")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) UpdateFreeFields(ctx context.Context, eaid string, fields *freeFieldsRequest) error {
	log.Printf("Updating account %s free fields %v\n", eaid, *fields)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(fields).
		SetHeader(headerContentType, contentTypeEmlJson).
		Put("/3.0/accounts/{id}/freefields")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) Transfer(ctx context.Context, eaid string, request *TransferRequest) error {
	log.Printf("Performing account %s transfer %s to %s", eaid, request.Amount, request.DestinationAccountId)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(request).
		SetHeader(headerContentType, contentTypeEmlJson).
		Post("/3.0/accounts/{id}/transfer")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) AddHook(ctx context.Context, request *HookRequest) (string, error) {
	log.Println("Adding notifications webhook", request.Uri, "scope", request.Scope)
	resp, err := e.request(ctx).
		SetBody(request).
		SetResult(&IdModel{}).
		SetHeader(headerContentType, contentTypeJson).
		SetHeader(headerAccept, contentTypeJson).
		Post("/3.0/hooks")
	if err := checkError(resp, err); err != nil {
		return "", err
	}
	return resp.Result().(*IdModel).Id, nil
}

func (e *emlStore) GetHooks(ctx context.Context) (*HookPage, error) {
	log.Println("Getting notification webhooks")
	hooks := &[]Hook{}
	resp, err := e.request(ctx).
		SetResult(hooks).
		SetHeader(headerAccept, contentTypeJson).
		Get("/3.0/hooks")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	actualPageSize, _ := strconv.Atoi(resp.Header().Get(headerPageSize))
	totalItems, _ := strconv.Atoi(resp.Header().Get(headerTotalItems))

	return &HookPage{
		NextCursor: "",
		PageSize:   actualPageSize,
		Count:      totalItems,
		Items:      *hooks,
	}, nil
}

func (e *emlStore) GetHook(ctx context.Context, hookId string) (*Hook, error) {
	log.Println("Getting notification webhook", hookId)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": hookId}).
		SetResult(&Hook{}).
		SetHeader(headerAccept, contentTypeJson).
		Get("/3.0/hooks/{id}")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*Hook), nil
}

func (e *emlStore) DeleteHook(ctx context.Context, hookId string) error {
	log.Println("Deleting notification webhook", hookId)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": hookId}).
		SetHeader(headerAccept, contentTypeJson).
		Delete("/3.0/hooks/{id}")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) UpdateHookScope(ctx context.Context, hookId string, scope []int) error {
	log.Println("Updating notification webhook", hookId, "scope", scope)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": hookId}).
		SetBody(HookRequest{Scope: scope}).
		SetHeader(headerContentType, contentTypeJson).
		SetHeader(headerAccept, contentTypeJson).
		Patch("/3.0/hooks/{id}")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) GetUndeliverable(ctx context.Context, hookId string, pageSize int, pageNumber int) (*MessagePage, error) {
	log.Println("Getting notification webhook", hookId, "undeliverable messages")
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": hookId}).
		SetQueryParams(map[string]string{queryPageNumber: strconv.Itoa(pageNumber), queryPageSize: strconv.Itoa(pageSize)}).
		SetResult([]Message{}).
		SetHeader(headerAccept, contentTypeJson).
		Get("/3.0/hooks/{id}/undeliverable")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	actualPageSize, _ := strconv.Atoi(resp.Header().Get(headerPageSize))
	totalPages, _ := strconv.Atoi(resp.Header().Get(headerTotalPages))
	page := &MessagePage{
		PageSize: actualPageSize,
		More:     totalPages > pageNumber,
	}
	if resp.StatusCode() != http.StatusNoContent {
		page.Items = *resp.Result().(*[]Message)
	}
	return page, nil
}

func (e *emlStore) DismissUndeliverable(ctx context.Context, hookId string, messageIds []string) error {
	log.Println("Dismissing notification webhook", hookId, "messages", messageIds)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": hookId}).
		SetBody(&MessageIdsRequest{MessageIds: messageIds}).
		SetHeader(headerContentType, contentTypeJson).
		SetHeader(headerAccept, contentTypeJson).
		Post("/3.0/hooks/{id}/undeliverable/dismiss")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

func (e *emlStore) Authenticate(ctx context.Context, eaid string, req AuthenticateRequest) (*AuthenticateResponse, error) {
	log.Printf("Authenticating account %s from IP %s", eaid, req.IPAddress)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(req).
		SetResult(&AuthenticateResponse{}).
		SetHeader(headerContentType, contentTypeEmlJson).
		Post("/3.0/accounts/{id}/authenticate")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*AuthenticateResponse), nil
}

func (e *emlStore) Initiate(ctx context.Context, eaid string, req InitiateRequest) (*InitiateResponse, error) {
	log.Printf("Initiating %s operation for account %s via %s", req.OperationType, eaid, req.CommunicateMethod)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(req).
		SetResult(&InitiateResponse{}).
		SetHeader(headerContentType, contentTypeEmlJson).
		Post("/3.0/accounts/{id}/initiate")
	if err := checkError(resp, err); err != nil {
		return nil, err
	}
	return resp.Result().(*InitiateResponse), nil
}

func (e *emlStore) Activate(ctx context.Context, eaid string, req ActivateRequest) error {
	log.Printf("Activating operation %s for account %s", req.ValidationData.OperationID, eaid)
	resp, err := e.request(ctx).
		SetPathParams(map[string]string{"id": eaid}).
		SetBody(req).
		SetHeader(headerContentType, contentTypeEmlJson).
		Post("/3.0/accounts/{id}/activate")
	if err := checkError(resp, err); err != nil {
		return err
	}
	return nil
}

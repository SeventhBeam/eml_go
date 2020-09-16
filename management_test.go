package eml

//func TestProcessUndeliverable(t *testing.T) {
//	emlStore := &eml.MockStore{}
//	qService := &queues.MockService{}
//	input := &queues.UndeliverableRequest{HookId: "hookId", LastUndeliverable: "5"}
//	d := MessageData{CompanyId: 2}
//	page1 := &MessagePage{More: true, PageSize: 2, Items: []Message{{Id: "1", Data: d}, {Id: "2", Data: d}}}
//	page2 := &MessagePage{More: true, PageSize: 2, Items: []Message{{Id: "3", Data: d}, {Id: "4", Data: d}}}
//	page3 := &MessagePage{More: false, PageSize: 2, Items: []Message{{Id: "5", Data: d}}}
//	emlStore.On("GetUndeliverable", ctx, "hookId", 20, 1).Return(page1, nil)
//	emlStore.On("GetUndeliverable", ctx, "hookId", 2, 2).Return(page2, nil)
//	emlStore.On("GetUndeliverable", ctx, "hookId", 2, 3).Return(page3, nil)
//	qService.On("AddProcessTransaction", ctx, &queues.TransactionRequest{MessageId: "1", CompanyId: "2"}).Return(nil)
//	qService.On("AddProcessTransaction", ctx, &queues.TransactionRequest{MessageId: "2", CompanyId: "2"}).Return(nil)
//	qService.On("AddProcessTransaction", ctx, &queues.TransactionRequest{MessageId: "3", CompanyId: "2"}).Return(nil)
//	qService.On("AddProcessTransaction", ctx, &queues.TransactionRequest{MessageId: "4", CompanyId: "2"}).Return(nil)
//	qService.On("AddProcessTransaction", ctx, &queues.TransactionRequest{MessageId: "5", CompanyId: "2"}).Return(nil)
//	emlStore.On("DismissUndeliverable", ctx, "hookId", []string{"1", "2", "3", "4", "5"}).Return(nil)
//
//	err := ProcessUndeliverableAlert(ctx, input, emlStore, qService)
//	assert.NoError(t, err, "Unexpected error")
//	mock.AssertExpectationsForObjects(t, emlStore, qService)
//}

//func TestSetupHook_new(t *testing.T) {
//	dataStore := &data.MockStore{}
//	emlStore := &eml.MockStore{}
//	e := &env.Settings{FunctionHost: "http://function/"}
//	emlConfig := &data.EmlConfig{DisbursementCompanyId: "12345", ProductCompanies: []data.ProductCompany{{CompanyId: "67890"}}}
//	dataStore.On("GetEmlConfig", ctx).Return(emlConfig, nil)
//	var keyId, keySecret string
//	emlStore.On("AddHook", ctx, mock.MatchedBy(func(req *eml.HookRequest) bool {
//		keyId = req.HmacKeyId
//		keySecret = req.HmacKeySecret
//		return reflect.DeepEqual(req.Scope, []int{12345, 67890}) && *req.Enabled && req.Uri == "http://function/Webhook/v1/eml/notification"
//	})).Return("hookId", nil)
//	dataStore.On("UpdateEmlConfigWebhook", ctx, "hookId", mock.MatchedBy(func(key *models.Key) bool {
//		secretHex, _ := key.SecretHex()
//		return key.Id == keyId && reflect.DeepEqual(secretHex, keySecret)
//	})).Return(nil)
//
//	err := SetupHook(ctx, dataStore, emlStore, nil, e)
//	assert.NoError(t, err, "Unexpected error")
//	mock.AssertExpectationsForObjects(t, dataStore, emlStore)
//}
//
//func TestSetupHook_existingOk(t *testing.T) {
//	dataStore := &data.MockStore{}
//	emlStore := &eml.MockStore{}
//	e := &env.Settings{FunctionHost: "http://function/"}
//	emlConfig := &data.EmlConfig{NotificationHookId: "hookId", DisbursementCompanyId: "12345", ProductCompanies: []data.ProductCompany{{CompanyId: "67890"}}}
//	dataStore.On("GetEmlConfig", ctx).Return(emlConfig, nil)
//	emlStore.On("GetHook", ctx, "hookId").Return(&eml.Hook{Id: "hookId", Scope: []int{67890, 12345}}, nil)
//
//	err := SetupHook(ctx, dataStore, emlStore, nil, e)
//	assert.NoError(t, err, "Unexpected error")
//	mock.AssertExpectationsForObjects(t, dataStore, emlStore)
//}
//
//func TestSetupHook_existingUpdateScope(t *testing.T) {
//	dataStore := &data.MockStore{}
//	emlStore := &eml.MockStore{}
//	e := &env.Settings{FunctionHost: "http://function/"}
//	emlConfig := &data.EmlConfig{NotificationHookId: "hookId", DisbursementCompanyId: "12345", ProductCompanies: []data.ProductCompany{{CompanyId: "67890"}, {CompanyId: "24680"}}}
//	dataStore.On("GetEmlConfig", ctx).Return(emlConfig, nil)
//	emlStore.On("GetHook", ctx, "hookId").Return(&eml.Hook{Id: "hookId", Scope: []int{67890, 12345}}, nil)
//	emlStore.On("UpdateHookScope", ctx, "hookId", []int{12345, 24680, 67890}).Return(nil)
//
//	err := SetupHook(ctx, dataStore, emlStore, nil, e)
//	assert.NoError(t, err, "Unexpected error")
//	mock.AssertExpectationsForObjects(t, dataStore, emlStore)
//}
//
//func TestSetupHook_existingUndelivered(t *testing.T) {
//	dataStore := &data.MockStore{}
//	emlStore := &eml.MockStore{}
//	qService := &queues.MockService{}
//	e := &env.Settings{FunctionHost: "http://function/"}
//	emlConfig := &data.EmlConfig{NotificationHookId: "hookId", DisbursementCompanyId: "12345", ProductCompanies: []data.ProductCompany{{CompanyId: "67890"}}}
//	page1 := &eml.MessagePage{More: false, PageSize: 1, Items: []eml.Message{{Id: "1", Data: eml.MessageData{CompanyId: 2}}}}
//	dataStore.On("GetEmlConfig", ctx).Return(emlConfig, nil)
//	emlStore.On("GetHook", ctx, "hookId").Return(&eml.Hook{Id: "hookId", Scope: []int{67890, 12345}, LastUndeliverable: "1"}, nil)
//	emlStore.On("GetUndeliverable", ctx, "hookId", 20, 1).Return(page1, nil)
//	qService.On("AddProcessTransaction", ctx, &queues.TransactionRequest{MessageId: "1", CompanyId: "2"}).Return(nil)
//	emlStore.On("DismissUndeliverable", ctx, "hookId", []string{"1"}).Return(nil)
//
//	err := SetupHook(ctx, dataStore, emlStore, qService, e)
//	assert.NoError(t, err, "Unexpected error")
//	mock.AssertExpectationsForObjects(t, dataStore, emlStore, qService)
//}

package eml

import (
	"log"
	"sort"
	"strconv"
	"time"
)

func mapBearerToken(model *TokenResponse) *bearerToken {
	return &bearerToken{
		Value:   model.AccessToken,
		Expires: time.Now().Add(time.Second * time.Duration(model.ExpiresIn)),
	}
}

func mapRegistrationUpdate(account *AccountInfo, info RegistrationInfo) *UpdateAccountRequest {
	registration := account.Personal
	registration.EmailAddress = info.Email()
	registration.MobileNumber = info.Phone()
	registration.FirstName = info.FirstName()
	registration.LastName = info.LastName()
	return &UpdateAccountRequest{
		InitiatingUserId: info.Uid(),
		Registration:     registration,
		PortalIdentifier: account.PortalIdentifier,
		MdesConfigId:     account.MdesConfigId,
		AccountExpiry:    account.AccountExpiry,
	}
}

func mapHookRequest(s *Settings, emlConfig *Config, key *Key) (*HookRequest, error) {
	scope, err := mapScope(emlConfig)
	secretHex, err := key.SecretHex()
	if err != nil {
		return nil, ContextualError(err, "key.SecretHex")
	}
	enabled := true
	return &HookRequest{
		Uri:             s.HookUri,
		Scope:           scope,
		FilterSpec:      FilterSpecAll,
		Enabled:         &enabled,
		ReliabilityMode: StoreUndeliverable,
		HmacKeyId:       key.Id,
		HmacKeySecret:   secretHex,
	}, nil
}

func mapScope(emlConfig *Config) ([]int, error) {
	numProductCompanies := len(emlConfig.ProductCompanies)
	scope := make([]int, numProductCompanies)
	for i, company := range emlConfig.ProductCompanies {
		num, err := strconv.Atoi(company.CompanyId)
		if err != nil {
			return nil, ContextualError(err, "error converting product company ID to int: %s", company.CompanyId)
		}
		scope[i] = num
	}
	num, err := strconv.Atoi(emlConfig.DisbursementCompanyId)
	if err == nil {
		scope = append(scope, num)
		log.Printf("error converting disbursement company ID to int: %s", emlConfig.DisbursementCompanyId)
	}
	sort.Ints(scope)
	return scope, nil
}

package eml

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

type TransactionHandler func(ctx context.Context, message *Message) error

// Webhook endpoint
func HandleNotification(ctx context.Context, req *Request, res Response, deps Dependencies, h TransactionHandler, uh TransactionHandler) error { //, deps *app.Dependencies
	messageType, version, err := GetEmlMessageSpec(req)
	log.Printf("Handling EML Notification %s@%s\n", messageType, version)
	if err != nil {
		return err
	}
	if messageType != TxnTypeTransaction && messageType != TxnTypeUndeliverableAlert {
		log.Println("Acknowledging ", messageType, " we don't care about")
		var message IdModel
		if err := json.NewDecoder(req.Body).Decode(message); err != nil {
			return BadError(ErrorParsingBody, ContextualError(err, "json.Decode"))
		}

		res.JsonOk(message)
		return nil
	}
	if !strings.HasPrefix(version, "1.") {
		return NotImplementedError(fmt.Errorf("message specification %s version %s is not supported", messageType, version))
	}
	//if emlConfig == nil || emlConfig.NotificationHookId == "" {
	//	emlConfig, err = deps.Data.GetEmlConfig(ctx)
	//	if err != nil {
	//		return ContextualError(err, "deps.Data.GetEmlConfig")
	//	}
	//}
	if len(deps.Config.HmacKeys) == 0 {
		return fmt.Errorf("no HMAC keys configured")
	}
	var message Message
	var buf bytes.Buffer
	if err := req.UnmarshalJsonAndCopy(&message, &buf); err != nil {
		return err
	}
	if message.HookId != deps.Config.NotificationHookId {
		// Would happen on dev/staging using same EML env and companies
		log.Println("Acknowledging", messageType, "meant for different hook", message.HookId, ", not ours:", deps.Config.NotificationHookId)
		res.JsonOk(IdModel{Id: message.Id})
		return nil
	}

	// Temporarily log entire payload
	var buf2 bytes.Buffer
	sb := new(strings.Builder)
	_, _ = io.Copy(sb, io.TeeReader(&buf, &buf2))
	fmt.Println(sb.String())
	buf = buf2
	// end temporary logging

	if body, err := req.CheckHmacSignature(deps.Config.HmacKeys, &buf); err != nil {
		log.Println(body.(*bytes.Buffer).String())
		return UnauthorizedError(ContextualError(err, "req.CheckHmacSignature"))
	}
	if messageType == TxnTypeUndeliverableAlert {
		if uh != nil {
			err = uh(ctx, &message)
		} else {
			err = processAllUndeliverableMessages(ctx, message.HookId, message.Id, deps.Store, h)
		}
	} else {
		err = h(ctx, &message)
	}
	return res.HandleJsonOk(IdModel{Id: message.Id}, err)
}

func GetEmlMessageSpec(r *Request) (messageType TransactionType, version string, err error) {
	spec := r.Header.Get(headerEmlSpecification)
	if spec == "" || !strings.Contains(spec, "@") {
		err = BadError(ErrorInvalidHeader.Format(headerEmlSpecification), fmt.Errorf("invalid specification header: %s", spec))
	} else {
		parts := strings.Split(spec, "@")
		messageType, version = TransactionType(parts[0]), parts[1]
	}
	return
}

type Config struct {
	ProductCompanies      []ProductCompany
	DisbursementCompanyId string
	NotificationHookId    string
	HmacKeys              []Key
}

type ProductCompany struct {
	CompanyId    string `firestore:"companyId"`
	IsPlastic    bool   `firestore:"isPlastic"`
	IsReloadable bool   `firestore:"isReloadable"`
}

type Key struct {
	Id     string `json:"id" firestore:"id"`
	Secret string `json:"secret" firestore:"secret"`
}

func (k *Key) SecretBase64() string {
	return k.Secret
}

func (k *Key) SecretBytes() ([]byte, error) {
	return base64.StdEncoding.DecodeString(k.Secret)
}

func (k *Key) SecretHex() (string, error) {
	b, err := k.SecretBytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

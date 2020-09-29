package eml

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"sort"
)

func SetupHook(ctx context.Context, emlConfig *Config, emlStore Store, s *Settings, h TransactionHandler) error {
	hooks, err := emlStore.GetHooks(ctx)
	//Get more hooks if there is a page cursor
	if err != nil {
		return ContextualError(err, "emlStore.GetHooks")
	}

	myHooks := make([]Hook, 0)
	for _, v := range hooks.Items {
		if v.HmacKeyId == emlConfig.NotificationHookId && v.Uri == s.hookUri() {
			myHooks = append(myHooks, v)
		}
	}

	if len(myHooks) == 0 {
		return registerNewHook(ctx, emlConfig, emlStore, s)
	}

	err = checkHookStatus(ctx, emlConfig, emlStore, h)
	if err != nil {
		return err
	}

	scope := myHooks[0].Scope
	for _, hook := range myHooks[1:] {
		scope = append(scope, hook.Scope...)
		//Remove the extra hooks, they will be covered by the updated scope
		err = emlStore.DeleteHook(ctx, hook.Id)
		if err != nil {
			log.Printf("Unable to delete hook: %v, due to %v", hook.Id, err)
		}
	}

	return nil
}

func registerNewHook(ctx context.Context, emlConfig *Config, emlStore Store, s *Settings) error {
	log.Println("Registering new EML webhook")
	key, err := GenerateSecureKey(32)
	if err != nil {
		return ContextualError(err, "utils.GenerateSecureKey")
	}
	newHookRequest, err := mapHookRequest(s, emlConfig, key)
	if err != nil {
		return ContextualError(err, "mapHookRequest")
	}
	_, err = emlStore.AddHook(ctx, newHookRequest)
	if err != nil {
		return ContextualError(err, "emlStore.AddHook")
	}
	//if err := emlStore.UpdateEmlConfigWebhook(ctx, hookId, key); err != nil {
	//	return ContextualError(err, "dataStore.UpdateEmlConfigWebhook")
	//}
	return nil
}

func checkHookStatus(ctx context.Context, emlConfig *Config, emlStore Store, h TransactionHandler) error {
	log.Println("Checking status of existing EML webhook", emlConfig.NotificationHookId)
	hook, err := emlStore.GetHook(ctx, emlConfig.NotificationHookId)
	if err != nil {
		return ContextualError(err, "emlStore.GetHook")
	}
	requiredScope, err := mapScope(emlConfig)
	if err != nil {
		return ContextualError(err, "mapScope")
	}
	sort.Ints(hook.Scope)
	if !reflect.DeepEqual(requiredScope, hook.Scope) {
		log.Println("Hook", hook.Id, "scopes don't match,", hook.Scope, "required:", requiredScope)
		if err := emlStore.UpdateHookScope(ctx, hook.Id, requiredScope); err != nil {
			return ContextualError(err, "emlStore.UpdateHookScope")
		}
	}
	if hook.LastUndeliverable != "" {
		log.Println("Hook", hook.Id, "has undelivered messages", hook.LastUndeliverable, hook.LastUndeliverableTimestamp)
		return processAllUndeliverableMessages(ctx, hook.Id, hook.LastUndeliverable, emlStore, h)
	}
	return nil
}

func processAllUndeliverableMessages(ctx context.Context, hookId, lastMessageId string, emlStore Store, h TransactionHandler) error {
	messages := collectMessages(ctx, hookId, emlStore)
	resolvedIds := handleMessages(ctx, messages, h)
	idList := make([]string, 0)
	handledLastMessage := false
	for id := range resolvedIds {
		idList = append(idList, id)
		if id == lastMessageId {
			handledLastMessage = true
		}
	}
	if len(idList) > 0 {
		err := emlStore.DismissUndeliverable(ctx, hookId, idList)
		if err != nil {
			log.Println("Error dismissing messages IDs:", err)
		}
	}
	if !handledLastMessage {
		return fmt.Errorf("last message with ID %s was not handled", lastMessageId)
	}
	return nil
}

func collectMessages(ctx context.Context, hookId string, emlStore Store) <-chan Message {
	out := make(chan Message)
	go func() {
		defer func() {
			close(out)
		}()
		page := &MessagePage{More: true, PageSize: 20}
		var err error
		for pageNumber := 1; page.More; pageNumber++ {
			page, err = emlStore.GetUndeliverable(ctx, hookId, page.PageSize, pageNumber)
			if err != nil {
				log.Println("emlStore.GetUndeliverable:", err)
				return
			}
			log.Println("Got page", pageNumber, "of undelivered messages:", len(page.Items), "more:", page.More)
			for _, message := range page.Items {
				select {
				case out <- message:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func handleMessages(ctx context.Context, messages <-chan Message, h TransactionHandler) <-chan string {
	out := make(chan string)
	go func() {
		defer func() {
			close(out)
		}()
		for message := range messages {
			err := h(ctx, &message)
			if err == nil {
				select {
				case out <- message.Id:
				case <-ctx.Done():
					return
				}
			} else {
				log.Println("error handling undelivered transaction", err)
			}
		}
	}()
	return out
}

func GenerateSecureKey(numBytes int) (*Key, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return &Key{
		Id:     UniqueID(),
		Secret: base64.StdEncoding.EncodeToString(b),
	}, nil
}

func UniqueID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

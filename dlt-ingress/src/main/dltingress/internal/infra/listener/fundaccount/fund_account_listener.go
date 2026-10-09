package fundaccountlistener

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/service/fundaccount"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
)

type Listener struct {
	appService fundaccount.AppServiceInterface
}

func NewListener(appService fundaccount.AppServiceInterface) *Listener {
	return &Listener{appService: appService}
}

func (l *Listener) Listen(ctx context.Context, evt *custodykey.KeyCreatedEvent) error {
	return l.appService.Execute(ctx, &fundaccount.Request{
		DltAccountId: evt.DltAccountId,
	})
}

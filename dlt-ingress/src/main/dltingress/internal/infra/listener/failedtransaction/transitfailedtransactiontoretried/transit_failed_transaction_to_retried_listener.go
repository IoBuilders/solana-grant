package transitfailedtransactiontoretriedlistener

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/service/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
)

type Listener struct {
	appService transitfailedtransactiontoretried.AppServiceInterface
}

func NewListener(appService transitfailedtransactiontoretried.AppServiceInterface) *Listener {
	return &Listener{appService: appService}
}

func (l *Listener) Listen(ctx context.Context, event *transaction.RetriedEvent) error {
	return l.appService.Execute(ctx, &transitfailedtransactiontoretried.Request{
		TxId: event.OldTxId,
	})
}

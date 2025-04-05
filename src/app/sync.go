package app

import (
	"sync"

	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
)

func RunSync() {
	environments.SetupEnvironments()

	sqlConnection := database.NewConnection()
	ioc.SetupDependencyInjection(sqlConnection)

	var wg sync.WaitGroup
	wg.Add(5)

	go ioc.UpdateTransactionBill.Execute()
	go ioc.UpdateUseCase.Execute("B4351E7E-F9AC-4A84-A113-A0E159303281")
	go ioc.UpdateUseCase.Execute("FF8C5393-2C43-4AE4-92F7-42AF4DD3AF08")
	go ioc.UpdateUseCase.Execute("45DE5288-D5D0-471A-BF18-09FE1FD2FC86")
	go ioc.UpdateUseCase.Execute("4FAE4733-FB19-4F0C-A678-3C6B7588F750")

	wg.Wait()
}

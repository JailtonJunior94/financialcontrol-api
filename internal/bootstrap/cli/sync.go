package cli

import (
	"sync"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/config"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
)

func RunSync() {
	config.SetupEnvironments()

	sqlConnection := database.NewConnection()
	container := bootstrapcontainer.Build(sqlConnection)

	var wg sync.WaitGroup
	wg.Add(5)

	go container.UpdateTransactionBill.Execute(&wg)
	go container.UpdateUseCase.Execute(&wg, "B4351E7E-F9AC-4A84-A113-A0E159303281")
	go container.UpdateUseCase.Execute(&wg, "FF8C5393-2C43-4AE4-92F7-42AF4DD3AF08")
	go container.UpdateUseCase.Execute(&wg, "45DE5288-D5D0-471A-BF18-09FE1FD2FC86")
	go container.UpdateUseCase.Execute(&wg, "4FAE4733-FB19-4F0C-A678-3C6B7588F750")

	wg.Wait()
}

package repository

import "ditto/pkg/domain"

type PrinterRepository interface {
	CreatePrinter(printer *domain.Printer) (error, *domain.Printer)
	UpdatePrinter(printerId string, printer *domain.Printer) (error, *domain.Printer)
	GetPrinterByExternalId(printerId string) (error, *domain.Printer)
	MultiGetPrintersByExternalId(printerIds []string) (error []*domain.Printer)
}

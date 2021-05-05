package repository

import (
	"context"
	"ditto/pkg/domain"
	"github.com/kutty-kumar/charminder/pkg"
	"github.com/kutty-kumar/ho_oh/core_v1"
)

type PrinterRepository interface {
	GetPrintersByUserId(ctx context.Context, userId string) ([]domain.Printer, error)
	DeletePrinter(ctx context.Context, userId string, printerId string) (*domain.Printer, error)
}

func NewPrinterGORMRepository(dao pkg.BaseDao) PrinterRepository {
	return &PrinterGORMRepository{
		dao,
	}
}

type PrinterGORMRepository struct {
	pkg.BaseDao
}

func (p *PrinterGORMRepository) GetPrintersByUserId(ctx context.Context, userId string) ([]domain.Printer, error) {
	var printers []domain.Printer
	if err := p.GetDb().WithContext(ctx).Table("printers").Where("user_id = ? AND status = 1", userId).Scan(&printers).Error; err != nil {
		return nil, err
	}
	return printers, nil
}

func (p *PrinterGORMRepository) DeletePrinter(ctx context.Context, userId string, printerId string) (*domain.Printer, error) {
	printer := &domain.Printer{}
	if err := p.GetDb().WithContext(ctx).Model(printer).Where("external_id = ? AND user_id = ?", printerId, userId).Find(printer).Error; err != nil {
		return nil, err
	}
	printer.Status = int(core_v1.Status_inactive)
	err, updatedPrinter := p.Update(ctx, printerId, printer)
	if err != nil {
		return nil, err
	}
	return updatedPrinter.(*domain.Printer), nil
}

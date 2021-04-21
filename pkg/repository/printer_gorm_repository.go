package repository

import "github.com/kutty-kumar/charminder/pkg"

type PrinterGORMRepository struct {
	pkg.BaseDao
}


func NewPrinterGORMRepository(baseDao pkg.BaseDao) *PrinterGORMRepository{
	return &PrinterGORMRepository{
		baseDao,
	}
}
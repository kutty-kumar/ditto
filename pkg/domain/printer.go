package domain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/kutty-kumar/charminder/pkg"
	"github.com/kutty-kumar/ho_oh/core_v1"
	ditto "github.com/kutty-kumar/ho_oh/ditto_v1"
)

type Printer struct {
	pkg.BaseDomain
	Name          string
	UserId        string
	SerialNumber  string
	ProductNumber string
	Description   string
	Status        int
}

func (p *Printer) MarshalBinary() ([]byte, error) {
	dto := p.ToDto().(ditto.PrinterDto)
	printerBytes, err := proto.Marshal(&dto)
	if err != nil {
		return nil, err
	}
	return printerBytes, nil
}

func (p *Printer) UnmarshalBinary(buffer []byte) error {
	dto := ditto.PrinterDto{}
	err := proto.Unmarshal(buffer, &dto)
	if err != nil {
		return err
	}
	p.FillProperties(&dto)
	return nil
}

func (p *Printer) GetName() pkg.DomainName {
	return "printers"
}

func (p *Printer) ToDto() interface{} {
	return ditto.PrinterDto{
		ExternalId:    p.ExternalId,
		Name:          p.Name,
		Description:   p.Description,
		SerialNumber:  p.SerialNumber,
		ProductNumber: p.ProductNumber,
		Status:        core_v1.Status(p.Status),
	}
}

func (p *Printer) FillProperties(dto interface{}) pkg.Base {
	printerDto := dto.(*ditto.PrinterDto)
	p.Name = printerDto.Name
	p.Description = printerDto.Description
	p.SerialNumber = printerDto.SerialNumber
	p.ProductNumber = printerDto.ProductNumber
	p.Status = int(printerDto.Status)
	return p
}

func (p *Printer) Merge(other interface{}) {
	otherPrinter := other.(*Printer)
	if otherPrinter.Name != "" {
		p.Name = otherPrinter.Name
	}
	if otherPrinter.Description != "" {
		p.Description = otherPrinter.Description
	}
}

func (p *Printer) FromSqlRow(rows *sql.Rows) (pkg.Base, error) {
	err := rows.Scan(&p.ExternalId, &p.Id, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &p.Status, &p.Name, &p.UserId, &p.SerialNumber, &p.ProductNumber,  &p.Description)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Printer) SetExternalId(externalId string) {
	p.ExternalId = externalId
}

func (p *Printer) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (p *Printer) String() string {
	return fmt.Sprintf("{\"name\": \"%v\",\"description\": \"%v\", \"serial_number\":\"%v\", \"product_number\": \"%v\"}", p.Name, p.Description, p.SerialNumber, p.ProductNumber)
}

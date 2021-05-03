package svc

import (
	"context"
	"ditto/pkg/domain"
	"ditto/pkg/repository"
	"github.com/kutty-kumar/charminder/pkg"
	ditto "github.com/kutty-kumar/ho_oh/ditto_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PrinterSvc struct {
	pkg.BaseSvc
	Repository repository.PrinterRepository
}

func NewPrinterSvc(baseSvc *pkg.BaseSvc, repository repository.PrinterRepository) *PrinterSvc {
	return &PrinterSvc{
		*baseSvc,
		repository,
	}
}

func (p *PrinterSvc) ToDto(printer *domain.Printer) ditto.PrinterDto {
	printerDto := printer.ToDto()
	return printerDto.(ditto.PrinterDto)
}

func (p *PrinterSvc) CreatePrinter(ctx context.Context, request *ditto.CreatePrinterRequest) (*ditto.CreatePrinterResponse, error) {
	printer := domain.Printer{}
	printer.FillProperties(request.Request)
	err, cPrinter := p.Create(ctx, &printer)
	if err != nil {
		return nil, err
	}
	dto := p.ToDto(cPrinter.(*domain.Printer))
	return &ditto.CreatePrinterResponse{Response: &dto}, nil
}

func (p *PrinterSvc) UpdatePrinter(ctx context.Context, request *ditto.UpdatePrinterRequest) (*ditto.UpdatePrinterResponse, error) {
	updatedPrinter := domain.Printer{}
	updatedPrinter.FillProperties(request.Request)
	err, uPrinter := p.Update(ctx, request.PrinterId, &updatedPrinter)
	if err != nil {
		return nil, err
	}
	dto := p.ToDto(uPrinter.(*domain.Printer))
	return &ditto.UpdatePrinterResponse{Response: &dto}, nil
}

func (p *PrinterSvc) GetPrinterByExternalId(ctx context.Context, request *ditto.GetPrinterByExternalIdRequest) (*ditto.GetPrinterByExternalIdResponse, error) {
	err, printer := p.FindByExternalId(ctx, request.PrinterId)
	if err != nil {
		return nil, err
	}
	dto := p.ToDto(printer.(*domain.Printer))
	return &ditto.GetPrinterByExternalIdResponse{Response: &dto}, nil
}

func (p *PrinterSvc) MultiGetPrintersByExternalId(ctx context.Context, request *ditto.MultiGetPrintersByExternalIdRequest) (*ditto.MultiGetPrintersByExternalIdResponse, error) {
	var dtoResponse []*ditto.PrinterDto
	err, printers := p.MultiGetByExternalId(ctx, request.PrinterIds)
	if err != nil {
		return nil, err
	}
	for _, printer := range printers {
		dto := p.ToDto(printer.(*domain.Printer))
		dtoResponse = append(dtoResponse, &dto)
	}
	return &ditto.MultiGetPrintersByExternalIdResponse{Result: dtoResponse}, nil
}

func (p *PrinterSvc) MultiGetPrintersForUser(ctx context.Context, req *ditto.NoOpRequest) (*ditto.MultiGetPrintersByExternalIdResponse, error) {
	userId := ctx.Value("user_id").(string)
	if len(userId) > 0 {
		printers, err := p.Repository.GetPrintersByUserId(ctx, userId)
		if err != nil {
			return nil, err
		}
		var result []*ditto.PrinterDto
		for _, printer := range printers {
			dto := printer.ToDto().(ditto.PrinterDto)
			result = append(result, &dto)
		}
		return &ditto.MultiGetPrintersByExternalIdResponse{Result: result}, nil
	}
	return nil, status.Errorf(codes.NotFound, "printers not found for user %v", userId)
}

func (p *PrinterSvc) DeletePrinter(ctx context.Context, req *ditto.DeletePrinterRequest) (*ditto.UpdatePrinterResponse, error) {
	userId := ctx.Value("user_id").(string)
	if len(userId) > 0 {
		updatedPrinter, err := p.Repository.DeletePrinter(ctx, userId, req.PrinterId)
		if err != nil {
			return nil, err
		}
		dto := updatedPrinter.ToDto().(ditto.PrinterDto)
		return &ditto.UpdatePrinterResponse{Response: &dto}, nil
	}
	return nil, status.Errorf(codes.NotFound, "printer not found for user %v", userId)
}

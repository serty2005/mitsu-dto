package driver

import (
	"mitsuscanner/pkg/mitsudriver"
	"mitsuscanner/internal/domain/models"
)

// ConvertDomainToDriverFiscalInfo преобразует доменную модель FiscalInfo в тип mitsudriver.FiscalInfo.
func ConvertDomainToDriverFiscalInfo(domain *models.FiscalInfo) *mitsudriver.FiscalInfo {
	if domain == nil {
		return nil
	}
	return &mitsudriver.FiscalInfo{
		ModelName:        domain.ModelName,
		SerialNumber:     domain.SerialNumber,
		RNM:              domain.RNM,
		OrganizationName: domain.OrganizationName,
		Address:          domain.Address,
		Inn:              domain.Inn,
		FnSerial:         domain.FnSerial,
		RegistrationDate: domain.RegistrationDate,
		FdNumber:         domain.FdNumber,
		FnEndDate:        domain.FnEndDate,
		OfdName:          domain.OfdName,
		SoftwareDate:     domain.SoftwareDate,
		FfdVersion:       domain.FfdVersion,
		FnExecution:      domain.FnExecution,
		FnEdition:        domain.FnEdition,
		AttributeExcise:  domain.AttributeExcise,
		AttributeMarked:  domain.AttributeMarked,
	}
}

// ConvertDriverToDomainFiscalInfo преобразует mitsudriver.FiscalInfo в доменную модель models.FiscalInfo.
func ConvertDriverToDomainFiscalInfo(driverInfo *mitsudriver.FiscalInfo) *models.FiscalInfo {
	if driverInfo == nil {
		return nil
	}
	return &models.FiscalInfo{
		ModelName:        driverInfo.ModelName,
		SerialNumber:     driverInfo.SerialNumber,
		RNM:              driverInfo.RNM,
		OrganizationName: driverInfo.OrganizationName,
		Address:          driverInfo.Address,
		Inn:              driverInfo.Inn,
		FnSerial:         driverInfo.FnSerial,
		RegistrationDate: driverInfo.RegistrationDate,
		FdNumber:         driverInfo.FdNumber,
		FnEndDate:        driverInfo.FnEndDate,
		OfdName:          driverInfo.OfdName,
		SoftwareDate:     driverInfo.SoftwareDate,
		FfdVersion:       driverInfo.FfdVersion,
		FnExecution:      driverInfo.FnExecution,
		FnEdition:        driverInfo.FnEdition,
		AttributeExcise:  driverInfo.AttributeExcise,
		AttributeMarked:  driverInfo.AttributeMarked,
	}
}

// ConvertDomainToDriverPrinterSettings преобразует доменную модель PrinterSettings в тип mitsudriver.PrinterSettings.
func ConvertDomainToDriverPrinterSettings(domain models.PrinterSettings) mitsudriver.PrinterSettings {
	return mitsudriver.PrinterSettings{
		Model:    domain.Model,
		BaudRate: domain.BaudRate,
		Paper:    domain.Paper,
		Font:     domain.Font,
		Width:    domain.Width,
		Length:   domain.Length,
	}
}

// ConvertDriverToDomainPrinterSettings преобразует mitsudriver.PrinterSettings в доменную модель models.PrinterSettings.
func ConvertDriverToDomainPrinterSettings(driverSettings mitsudriver.PrinterSettings) models.PrinterSettings {
	return models.PrinterSettings{
		Model:    driverSettings.Model,
		BaudRate: driverSettings.BaudRate,
		Paper:    driverSettings.Paper,
		Font:     driverSettings.Font,
		Width:    driverSettings.Width,
		Length:   driverSettings.Length,
	}
}

// ConvertDomainToDriverDrawerSettings преобразует доменную модель DrawerSettings в тип mitsudriver.DrawerSettings.
func ConvertDomainToDriverDrawerSettings(domain models.DrawerSettings) mitsudriver.DrawerSettings {
	return mitsudriver.DrawerSettings{
		Pin:  domain.Pin,
		Rise: domain.Rise,
		Fall: domain.Fall,
	}
}

// ConvertDriverToDomainDrawerSettings преобразует mitsudriver.DrawerSettings в доменную модель models.DrawerSettings.
func ConvertDriverToDomainDrawerSettings(driverSettings mitsudriver.DrawerSettings) models.DrawerSettings {
	return models.DrawerSettings{
		Pin:  driverSettings.Pin,
		Rise: driverSettings.Rise,
		Fall: driverSettings.Fall,
	}
}

// ConvertDomainToDriverClicheLineData преобразует доменную модель ClicheLineData в тип mitsudriver.ClicheLineData.
func ConvertDomainToDriverClicheLineData(domain []models.ClicheLineData) []mitsudriver.ClicheLineData {
	result := make([]mitsudriver.ClicheLineData, len(domain))
	for i, line := range domain {
		result[i] = mitsudriver.ClicheLineData{
			Text:   line.Text,
			Format: line.Format,
		}
	}
	return result
}

// ConvertDriverToDomainClicheLineData преобразует mitsudriver.ClicheLineData в доменную модель models.ClicheLineData.
func ConvertDriverToDomainClicheLineData(driverData []mitsudriver.ClicheLineData) []models.ClicheLineData {
	result := make([]models.ClicheLineData, len(driverData))
	for i, line := range driverData {
		result[i] = models.ClicheLineData{
			Text:   line.Text,
			Format: line.Format,
		}
	}
	return result
}

// ConvertDomainToDriverLanSettings преобразует доменную модель LanSettings в тип mitsudriver.LanSettings.
func ConvertDomainToDriverLanSettings(domain models.LanSettings) mitsudriver.LanSettings {
	return mitsudriver.LanSettings{
		Addr: domain.Addr,
		Port: domain.Port,
		Mask: domain.Mask,
		Dns:  domain.Dns,
		Gw:   domain.Gw,
	}
}

// ConvertDriverToDomainLanSettings преобразует mitsudriver.LanSettings в доменную модель models.LanSettings.
func ConvertDriverToDomainLanSettings(driverSettings mitsudriver.LanSettings) models.LanSettings {
	return models.LanSettings{
		Addr: driverSettings.Addr,
		Port: driverSettings.Port,
		Mask: driverSettings.Mask,
		Dns:  driverSettings.Dns,
		Gw:   driverSettings.Gw,
	}
}

// ConvertDomainToDriverOfdSettings преобразует доменную модель OfdSettings в тип mitsudriver.OfdSettings.
func ConvertDomainToDriverOfdSettings(domain models.OfdSettings) mitsudriver.OfdSettings {
	return mitsudriver.OfdSettings{
		Addr:     domain.Addr,
		Port:     domain.Port,
		Client:   domain.Client,
		TimerFN:  domain.TimerFN,
		TimerOFD: domain.TimerOFD,
	}
}

// ConvertDriverToDomainOfdSettings преобразует mitsudriver.OfdSettings в доменную модель models.OfdSettings.
func ConvertDriverToDomainOfdSettings(driverSettings mitsudriver.OfdSettings) models.OfdSettings {
	return models.OfdSettings{
		Addr:     driverSettings.Addr,
		Port:     driverSettings.Port,
		Client:   driverSettings.Client,
		TimerFN:  driverSettings.TimerFN,
		TimerOFD: driverSettings.TimerOFD,
	}
}

// ConvertDomainToDriverOismSettings преобразует доменную модель ServerSettings (OISM) в тип mitsudriver.OismSettings.
func ConvertDomainToDriverOismSettings(domain models.OismSettings) mitsudriver.OismSettings {
	return mitsudriver.OismSettings{
		Addr: domain.Addr,
		Port: domain.Port,
	}
}

// ConvertDriverToDomainOismSettings преобразует mitsudriver.OismSettings в доменную модель models.ServerSettings.
func ConvertDriverToDomainOismSettings(driverSettings mitsudriver.OismSettings) models.OismSettings {
	return models.OismSettings{
		Addr: driverSettings.Addr,
		Port: driverSettings.Port,
	}
}

// ConvertDomainToDriverOkpSettings преобразует доменную модель ServerSettings (OKP) в тип mitsudriver.ServerSettings.
func ConvertDomainToDriverOkpSettings(domain models.ServerSettings) mitsudriver.ServerSettings {
	return mitsudriver.ServerSettings{
		Addr: domain.Addr,
		Okp:  domain.Okp,
		Port: domain.Port,
	}
}

// ConvertDriverToDomainOkpSettings преобразует mitsudriver.ServerSettings в доменную модель models.ServerSettings.
func ConvertDriverToDomainOkpSettings(driverSettings mitsudriver.ServerSettings) models.ServerSettings {
	return models.ServerSettings{
		Addr: driverSettings.Addr,
		Okp:  driverSettings.Okp,
		Port: driverSettings.Port,
	}
}

// ConvertDomainToDriverTaxRates преобразует доменную модель TaxRates в тип mitsudriver.TaxRates.
func ConvertDomainToDriverTaxRates(domain *models.TaxRates) *mitsudriver.TaxRates {
	if domain == nil {
		return nil
	}
	return &mitsudriver.TaxRates{
		T1:  domain.T1,
		T2:  domain.T2,
		T3:  domain.T3,
		T4:  domain.T4,
		T5:  domain.T5,
		T6:  domain.T6,
		T7:  domain.T7,
		T8:  domain.T8,
		T9:  domain.T9,
		T10: domain.T10,
	}
}

// ConvertDriverToDomainTaxRates преобразует mitsudriver.TaxRates в доменную модель models.TaxRates.
func ConvertDriverToDomainTaxRates(driverRates *mitsudriver.TaxRates) *models.TaxRates {
	if driverRates == nil {
		return nil
	}
	return &models.TaxRates{
		T1:  driverRates.T1,
		T2:  driverRates.T2,
		T3:  driverRates.T3,
		T4:  driverRates.T4,
		T5:  driverRates.T5,
		T6:  driverRates.T6,
		T7:  driverRates.T7,
		T8:  driverRates.T8,
		T9:  driverRates.T9,
		T10: driverRates.T10,
	}
}

// ConvertDomainToDriverRegData преобразует доменную модель RegData в тип mitsudriver.RegData.
func ConvertDomainToDriverRegData(domain *models.RegData) *mitsudriver.RegData {
	if domain == nil {
		return nil
	}
	return &mitsudriver.RegData{
		RNM:           domain.RNM,
		Inn:           domain.Inn,
		FfdVer:        domain.FfdVer,
		RegDate:       domain.RegDate,
		RegTime:       domain.RegTime,
		RegNumber:     domain.RegNumber,
		FdNumber:      domain.FdNumber,
		FpNumber:      domain.FpNumber,
		Base:          domain.Base,
		TaxSystems:    domain.TaxSystems,
		TaxBase:       domain.TaxBase,
		ModeMask:      domain.ModeMask,
		ExtModeMask:   domain.ExtModeMask,
		MarkAttr:      domain.MarkAttr,
		ExciseAttr:    domain.ExciseAttr,
		InternetAttr:  domain.InternetAttr,
		ServiceAttr:   domain.ServiceAttr,
		BsoAttr:       domain.BsoAttr,
		LotteryAttr:   domain.LotteryAttr,
		GamblingAttr:  domain.GamblingAttr,
		PawnAttr:      domain.PawnAttr,
		InsAttr:       domain.InsAttr,
		DineAttr:      domain.DineAttr,
		OptAttr:       domain.OptAttr,
		VendAttr:      domain.VendAttr,
		AutoModeAttr:  domain.AutoModeAttr,
		AutoNumAttr:   domain.AutoNumAttr,
		AutonomAttr:   domain.AutonomAttr,
		EncryptAttr:   domain.EncryptAttr,
		PrintAutoAttr: domain.PrintAutoAttr,
		OrgName:       domain.OrgName,
		Address:       domain.Address,
		Place:         domain.Place,
		OfdName:       domain.OfdName,
		OfdInn:        domain.OfdInn,
		Site:          domain.Site,
		EmailSender:   domain.EmailSender,
		AutoNumTag:    domain.AutoNumTag,
		FnSerial:      domain.FnSerial,
		FnEdition:     domain.FnEdition,
		PrinterSerial: domain.PrinterSerial,
	}
}

// ConvertDriverToDomainRegData преобразует mitsudriver.RegData в доменную модель models.RegData.
func ConvertDriverToDomainRegData(driverData *mitsudriver.RegData) *models.RegData {
	if driverData == nil {
		return nil
	}
	return &models.RegData{
		RNM:           driverData.RNM,
		Inn:           driverData.Inn,
		FfdVer:        driverData.FfdVer,
		RegDate:       driverData.RegDate,
		RegTime:       driverData.RegTime,
		RegNumber:     driverData.RegNumber,
		FdNumber:      driverData.FdNumber,
		FpNumber:      driverData.FpNumber,
		Base:          driverData.Base,
		TaxSystems:    driverData.TaxSystems,
		TaxBase:       driverData.TaxBase,
		ModeMask:      driverData.ModeMask,
		ExtModeMask:   driverData.ExtModeMask,
		MarkAttr:      driverData.MarkAttr,
		ExciseAttr:    driverData.ExciseAttr,
		InternetAttr:  driverData.InternetAttr,
		ServiceAttr:   driverData.ServiceAttr,
		BsoAttr:       driverData.BsoAttr,
		LotteryAttr:   driverData.LotteryAttr,
		GamblingAttr:  driverData.GamblingAttr,
		PawnAttr:      driverData.PawnAttr,
		InsAttr:       driverData.InsAttr,
		DineAttr:      driverData.DineAttr,
		OptAttr:       driverData.OptAttr,
		VendAttr:      driverData.VendAttr,
		AutoModeAttr:  driverData.AutoModeAttr,
		AutoNumAttr:   driverData.AutoNumAttr,
		AutonomAttr:   driverData.AutonomAttr,
		EncryptAttr:   driverData.EncryptAttr,
		PrintAutoAttr: driverData.PrintAutoAttr,
		OrgName:       driverData.OrgName,
		Address:       driverData.Address,
		Place:         driverData.Place,
		OfdName:       driverData.OfdName,
		OfdInn:        driverData.OfdInn,
		Site:          driverData.Site,
		EmailSender:   driverData.EmailSender,
		AutoNumTag:    driverData.AutoNumTag,
		FnSerial:      driverData.FnSerial,
		FnEdition:     driverData.FnEdition,
		PrinterSerial: driverData.PrinterSerial,
	}
}

// ConvertDomainToDriverShiftStatus преобразует доменную модель ShiftStatus в тип mitsudriver.ShiftStatus.
func ConvertDomainToDriverShiftStatus(domain *models.ShiftStatus) *mitsudriver.ShiftStatus {
	if domain == nil {
		return nil
	}
	return &mitsudriver.ShiftStatus{
		ShiftNum: domain.ShiftNum,
		State:    domain.State,
		Count:    domain.Count,
		FdNum:    domain.FdNum,
		KeyValid: domain.KeyValid,
		Ofd: struct {
			Count int    `xml:"COUNT,attr"`
			First int    `xml:"FIRST,attr"`
			Date  string `xml:"DATE,attr"`
			Time  string `xml:"TIME,attr"`
		}{
			Count: domain.Ofd.Count,
			First: domain.Ofd.First,
			Date:  domain.Ofd.Date,
			Time:  domain.Ofd.Time,
		},
	}
}

// ConvertDriverToDomainShiftStatus преобразует mitsudriver.ShiftStatus в доменную модель models.ShiftStatus.
func ConvertDriverToDomainShiftStatus(driverStatus *mitsudriver.ShiftStatus) *models.ShiftStatus {
	if driverStatus == nil {
		return nil
	}
	return &models.ShiftStatus{
		ShiftNum: driverStatus.ShiftNum,
		State:    driverStatus.State,
		Count:    driverStatus.Count,
		FdNum:    driverStatus.FdNum,
		KeyValid: driverStatus.KeyValid,
		Ofd: struct {
			Count int
			First int
			Date  string
			Time  string
		}{
			Count: driverStatus.Ofd.Count,
			First: driverStatus.Ofd.First,
			Date:  driverStatus.Ofd.Date,
			Time:  driverStatus.Ofd.Time,
		},
	}
}

// ConvertDomainToDriverShiftTotals преобразует доменную модель ShiftTotals в тип mitsudriver.ShiftTotals.
func ConvertDomainToDriverShiftTotals(domain *models.ShiftTotals) *mitsudriver.ShiftTotals {
	if domain == nil {
		return nil
	}
	return &mitsudriver.ShiftTotals{
		ShiftNum: domain.ShiftNum,
		Income: struct {
			Count string `xml:"COUNT,attr"`
			Total string `xml:"TOTAL,attr"`
		}{
			Count: domain.Income.Count,
			Total: domain.Income.Total,
		},
		Payout: struct {
			Count string `xml:"COUNT,attr"`
			Total string `xml:"TOTAL,attr"`
		}{
			Count: domain.Payout.Count,
			Total: domain.Payout.Total,
		},
		Cash: struct {
			Total string `xml:"TOTAL,attr"`
		}{
			Total: domain.Cash.Total,
		},
	}
}

// ConvertDriverToDomainShiftTotals преобразует mitsudriver.ShiftTotals в доменную модель models.ShiftTotals.
func ConvertDriverToDomainShiftTotals(driverTotals *mitsudriver.ShiftTotals) *models.ShiftTotals {
	if driverTotals == nil {
		return nil
	}
	return &models.ShiftTotals{
		ShiftNum: driverTotals.ShiftNum,
		Income: struct {
			Count string
			Total string
		}{
			Count: driverTotals.Income.Count,
			Total: driverTotals.Income.Total,
		},
		Payout: struct {
			Count string
			Total string
		}{
			Count: driverTotals.Payout.Count,
			Total: driverTotals.Payout.Total,
		},
		Cash: struct {
			Total string
		}{
			Total: driverTotals.Cash.Total,
		},
	}
}

// ConvertDomainToDriverFnStatus преобразует доменную модель FnStatus в тип mitsudriver.FnStatus.
func ConvertDomainToDriverFnStatus(domain *models.FnStatus) *mitsudriver.FnStatus {
	if domain == nil {
		return nil
	}
	return &mitsudriver.FnStatus{
		Serial:  domain.Serial,
		Ffd:     domain.Ffd,
		Phase:   domain.Phase,
		Valid:   domain.Valid,
		LastFD:  domain.LastFD,
		Flag:    domain.Flag,
		Edition: domain.Edition,
		Power:   domain.Power,
	}
}

// ConvertDriverToDomainFnStatus преобразует mitsudriver.FnStatus в доменную модель models.FnStatus.
func ConvertDriverToDomainFnStatus(driverStatus *mitsudriver.FnStatus) *models.FnStatus {
	if driverStatus == nil {
		return nil
	}
	return &models.FnStatus{
		Serial:  driverStatus.Serial,
		Ffd:     driverStatus.Ffd,
		Phase:   driverStatus.Phase,
		Valid:   driverStatus.Valid,
		LastFD:  driverStatus.LastFD,
		Flag:    driverStatus.Flag,
		Edition: driverStatus.Edition,
		Power:   driverStatus.Power,
	}
}

// ConvertDomainToDriverOfdExchangeStatus преобразует доменную модель OfdExchangeStatus в тип mitsudriver.OfdExchangeStatus.
func ConvertDomainToDriverOfdExchangeStatus(domain *models.OfdExchangeStatus) *mitsudriver.OfdExchangeStatus {
	if domain == nil {
		return nil
	}
	return &mitsudriver.OfdExchangeStatus{
		Count:    domain.Count,
		FirstDoc: domain.FirstDoc,
		Date:     domain.Date,
		Time:     domain.Time,
	}
}

// ConvertDriverToDomainOfdExchangeStatus преобразует mitsudriver.OfdExchangeStatus в доменную модель models.OfdExchangeStatus.
func ConvertDriverToDomainOfdExchangeStatus(driverStatus *mitsudriver.OfdExchangeStatus) *models.OfdExchangeStatus {
	if driverStatus == nil {
		return nil
	}
	return &models.OfdExchangeStatus{
		Count:    driverStatus.Count,
		FirstDoc: driverStatus.FirstDoc,
		Date:     driverStatus.Date,
		Time:     driverStatus.Time,
	}
}

// ConvertDomainToDriverMarkingStatus преобразует доменную модель MarkingStatus в тип mitsudriver.MarkingStatus.
func ConvertDomainToDriverMarkingStatus(domain *models.MarkingStatus) *mitsudriver.MarkingStatus {
	if domain == nil {
		return nil
	}
	return &mitsudriver.MarkingStatus{
		MarkState: domain.MarkState,
		Keep:      domain.Keep,
		Flag:      domain.Flag,
		Notice:    domain.Notice,
		Holds:     domain.Holds,
		Pending:   domain.Pending,
		Warning:   domain.Warning,
	}
}

// ConvertDriverToDomainMarkingStatus преобразует mitsudriver.MarkingStatus в доменную модель models.MarkingStatus.
func ConvertDriverToDomainMarkingStatus(driverStatus *mitsudriver.MarkingStatus) *models.MarkingStatus {
	if driverStatus == nil {
		return nil
	}
	return &models.MarkingStatus{
		MarkState: driverStatus.MarkState,
		Keep:      driverStatus.Keep,
		Flag:      driverStatus.Flag,
		Notice:    driverStatus.Notice,
		Holds:     driverStatus.Holds,
		Pending:   driverStatus.Pending,
		Warning:   driverStatus.Warning,
	}
}

// ConvertDomainToDriverRegResponse преобразует доменную модель RegResponse в тип mitsudriver.RegResponse.
func ConvertDomainToDriverRegResponse(domain *models.RegResponse) *mitsudriver.RegResponse {
	if domain == nil {
		return nil
	}
	return &mitsudriver.RegResponse{
		FdNumber: domain.FdNumber,
		FpNumber: domain.FpNumber,
	}
}

// ConvertDriverToDomainRegResponse преобразует mitsudriver.RegResponse в доменную модель models.RegResponse.
func ConvertDriverToDomainRegResponse(driverResponse *mitsudriver.RegResponse) *models.RegResponse {
	if driverResponse == nil {
		return nil
	}
	return &models.RegResponse{
		FdNumber: driverResponse.FdNumber,
		FpNumber: driverResponse.FpNumber,
	}
}

// ConvertDomainToDriverCloseFnResult преобразует доменную модель CloseFnResult в тип mitsudriver.CloseFnResult.
func ConvertDomainToDriverCloseFnResult(domain *models.CloseFnResult) *mitsudriver.CloseFnResult {
	if domain == nil {
		return nil
	}
	return &mitsudriver.CloseFnResult{
		FD: domain.FD,
		FP: domain.FP,
	}
}

// ConvertDriverToDomainCloseFnResult преобразует mitsudriver.CloseFnResult в доменную модель models.CloseFnResult.
func ConvertDriverToDomainCloseFnResult(driverResult *mitsudriver.CloseFnResult) *models.CloseFnResult {
	if driverResult == nil {
		return nil
	}
	return &models.CloseFnResult{
		FD: driverResult.FD,
		FP: driverResult.FP,
	}
}

// ConvertDomainToDriverRegistrationRequest преобразует доменную модель RegistrationRequest в тип mitsudriver.RegistrationRequest.
func ConvertDomainToDriverRegistrationRequest(domain models.RegistrationRequest) mitsudriver.RegistrationRequest {
	return mitsudriver.RegistrationRequest{
		IsReregistration: domain.IsReregistration,
		Base:             domain.Base,
		RNM:              domain.RNM,
		Inn:              domain.Inn,
		FfdVer:           domain.FfdVer,
		TaxSystems:       domain.TaxSystems,
		TaxSystemBase:    domain.TaxSystemBase,
		AutomatNumber:    domain.AutomatNumber,
		InternetCalc:     domain.InternetCalc,
		Service:          domain.Service,
		BSO:              domain.BSO,
		Lottery:          domain.Lottery,
		Gambling:         domain.Gambling,
		Excise:           domain.Excise,
		Marking:          domain.Marking,
		PawnShop:         domain.PawnShop,
		Insurance:        domain.Insurance,
		Catering:         domain.Catering,
		Wholesale:        domain.Wholesale,
		Vending:          domain.Vending,
		AutomatMode:      domain.AutomatMode,
		AutonomousMode:   domain.AutonomousMode,
		Encryption:       domain.Encryption,
		PrinterAutomat:   domain.PrinterAutomat,
		OrgName:          domain.OrgName,
		Address:          domain.Address,
		Place:            domain.Place,
		OfdName:          domain.OfdName,
		OfdInn:           domain.OfdInn,
		FnsSite:          domain.FnsSite,
		SenderEmail:      domain.SenderEmail,
	}
}

// ConvertDomainToDriverItemPosition преобразует доменную модель ItemPosition в тип mitsudriver.ItemPosition.
func ConvertDomainToDriverItemPosition(domain models.ItemPosition) mitsudriver.ItemPosition {
	return mitsudriver.ItemPosition{
		Name:     domain.Name,
		Price:    domain.Price,
		Quantity: domain.Quantity,
		Tax:      domain.Tax,
	}
}

// ConvertDomainToDriverPaymentInfo преобразует доменную модель PaymentInfo в тип mitsudriver.PaymentInfo.
func ConvertDomainToDriverPaymentInfo(domain models.PaymentInfo) mitsudriver.PaymentInfo {
	return mitsudriver.PaymentInfo{
		Type: domain.Type,
		Sum:  domain.Sum,
	}
}

// ConvertDomainToDriverDeviceOptions преобразует доменную модель DeviceOptions в тип mitsudriver.DeviceOptions.
func ConvertDomainToDriverDeviceOptions(domain models.DeviceOptions) mitsudriver.DeviceOptions {
	return mitsudriver.DeviceOptions{
		B0: domain.B0,
		B1: domain.B1,
		B2: domain.B2,
		B3: domain.B3,
		B4: domain.B4,
		B5: domain.B5,
		B6: domain.B6,
		B7: domain.B7,
		B8: domain.B8,
		B9: domain.B9,
	}
}

// ConvertDriverToDomainDeviceOptions преобразует mitsudriver.DeviceOptions в доменную модель models.DeviceOptions.
func ConvertDriverToDomainDeviceOptions(driverOptions mitsudriver.DeviceOptions) models.DeviceOptions {
	return models.DeviceOptions{
		B0: driverOptions.B0,
		B1: driverOptions.B1,
		B2: driverOptions.B2,
		B3: driverOptions.B3,
		B4: driverOptions.B4,
		B5: driverOptions.B5,
		B6: driverOptions.B6,
		B7: driverOptions.B7,
		B8: driverOptions.B8,
		B9: driverOptions.B9,
	}
}

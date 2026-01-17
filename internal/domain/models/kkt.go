package models

// FiscalInfo содержит агрегированную информацию о фискальном регистраторе.
type FiscalInfo struct {
	ModelName        string
	SerialNumber     string
	RNM              string
	OrganizationName string
	Address          string
	Inn              string
	FnSerial         string
	RegistrationDate string
	FdNumber         string
	FnEndDate        string
	OfdName          string
	SoftwareDate     string
	FfdVersion       string
	FnExecution      string
	FnEdition        string
	AttributeExcise  bool
	AttributeMarked  bool
}

// PrinterSettings содержит настройки принтера.
type PrinterSettings struct {
	Model    string
	BaudRate int
	Paper    int
	Font     int
	Width    int
	Length   int
}

// DrawerSettings содержит настройки денежного ящика.
type DrawerSettings struct {
	Pin  int
	Rise int
	Fall int
}

// ClicheLineData содержит данные одной строки клише.
type ClicheLineData struct {
	Text   string
	Format string // Строка вида "000000"
}

// LanSettings содержит настройки сети.
type LanSettings struct {
	Addr string
	Port int
	Mask string
	Dns  string
	Gw   string
}

// OfdSettings содержит настройки ОФД.
type OfdSettings struct {
	Addr     string
	Port     int
	Client   string
	TimerFN  int
	TimerOFD int
}

// OismSettings содержит настройки сервера OISM
type OismSettings struct {
	Addr string // Для OISM
	Port int
}

// ServerSettings содержит настройки сервера OKP.
type ServerSettings struct {
	Addr string
	Okp  string // Для OKP (имя атрибута отличается)
	Port int
}

// TaxRates содержит настройки налоговых ставок.
type TaxRates struct {
	T1  string // 20%
	T2  string // 10%
	T3  string // 20/120
	T4  string // 10/110
	T5  string // 0%
	T6  string // Без НДС
	T7  string // 5%
	T8  string // 7%
	T9  string // 5/105
	T10 string // 7/107
}

// RegData содержит полные данные о регистрации ККТ.
type RegData struct {
	RNM        string
	Inn        string
	FfdVer     string
	RegDate    string
	RegTime    string
	RegNumber  string // Порядковый номер регистрации
	FdNumber   string // Номер фискального документа
	FpNumber   string // Фискальный признак из T1077
	Base       string // Коды причин регистрации
	TaxSystems string // СНО (0,1,...)
	TaxBase    string

	// Маски режимов
	ModeMask    uint32
	ExtModeMask uint32

	// Флаги (атрибуты)
	MarkAttr      string // Маркировка
	ExciseAttr    string // Подакцизные
	InternetAttr  string // Интернет
	ServiceAttr   string // Услуги
	BsoAttr       string // БСО
	LotteryAttr   string // Лотерея
	GamblingAttr  string // Азартные
	PawnAttr      string // Ломбард
	InsAttr       string // Страхование
	DineAttr      string // Общепит
	OptAttr       string // Опт
	VendAttr      string // Вендинг
	AutoModeAttr  string // Автоматический режим
	AutoNumAttr   string // Номер автомата (атрибут)
	AutonomAttr   string // Автономный
	EncryptAttr   string // Шифрование
	PrintAutoAttr string // Принтер в автомате

	// Вложенные теги
	OrgName     string
	Address     string
	Place       string
	OfdName     string
	OfdInn      string
	Site        string
	EmailSender string
	AutoNumTag  string

	// Дополнительные поля для диалога
	FnSerial      string // Заводской номер ФН
	FnEdition     string // Исполнение ФН
	PrinterSerial string // Серийный номер ФР
}

// ShiftStatus содержит информацию о текущей смене.
type ShiftStatus struct {
	ShiftNum int
	State    string // 0-закрыта, 1-открыта, 9-истекла
	Count    int
	FdNum    int
	KeyValid int

	// Вложенная структура для статуса обмена ОФД
	Ofd struct {
		Count int
		First int
		Date  string
		Time  string
	}
}

// ShiftTotals содержит итоги смены.
type ShiftTotals struct {
	ShiftNum int
	Income   struct {
		Count string
		Total string
	}
	Payout struct {
		Count string
		Total string
	}
	Cash struct {
		Total string
	}
}

// FnStatus содержит информацию о фискальном накопителе.
type FnStatus struct {
	Serial  string
	Ffd     string
	Phase   string
	Valid   string
	LastFD  int
	Flag    string // HEX маска предупреждений
	Edition string
	Power   string // Флаг питания (1 = установлен, 0 = сброшен)
}

// OfdExchangeStatus содержит информацию об обмене с ОФД.
type OfdExchangeStatus struct {
	Count    int
	FirstDoc int
	Date     string
	Time     string
}

// MarkingStatus содержит информацию о маркировке.
type MarkingStatus struct {
	MarkState int
	Keep      int
	Flag      string
	Notice    int
	Holds     int
	Pending   int
	Warning   int
}

// RegResponse содержит ответ на команду регистрации/перерегистрации.
type RegResponse struct {
	FdNumber string // Номер фискального документа
	FpNumber string // Фискальный признак из T1077
}

// CloseFnResult содержит результат закрытия фискального архива.
type CloseFnResult struct {
	FD int    // номер фискального документа
	FP string // фискальный признак
}

// ReportFnCloseData содержит данные для отчета о закрытии фискального архива.
type ReportFnCloseData struct {
	DateTime  string
	FP        string
	FD        int
	RNM       string
	FNNumber  string
	KKTNumber string
	Address   string
	Place     string
}

// RegistrationRequest содержит параметры для регистрации ККТ.
type RegistrationRequest struct {
	// Атрибуты (Attributes)
	IsReregistration bool   // false = Регистрация, true = Перерегистрация
	Base             string // Коды причин перерегистрации (через запятую, напр "1,5")

	RNM            string // T1037 (Рег. номер)
	Inn            string // T1018 (ИНН Пользователя)
	FfdVer         string // T1209 (Версия ФФД)
	TaxSystems     string // T1062 (СНО через запятую: "0,1")
	TaxSystemBase  string // T1062_Base (Базовая СНО - опционально)
	AutomatNumber  string // T1036 (Номер автомата)
	InternetCalc   bool   // T1108 (Расчеты в Интернет)
	Service        bool   // T1109 (Услуги)
	BSO            bool   // T1110 (БСО)
	Lottery        bool   // T1126 (Лотерея)
	Gambling       bool   // T1193 (Азартные игры)
	Excise         bool   // T1207 (Подакцизные товары)
	Marking        bool   // MARK (Маркировка)
	PawnShop       bool   // PAWN (Ломбард)
	Insurance      bool   // INS (Страхование)
	Catering       bool   // DINE (Общепит)
	Wholesale      bool   // OPT (Опт)
	Vending        bool   // VEND (Вендинг)
	AutomatMode    bool   // T1001 (Автоматический режим)
	AutonomousMode bool   // T1002 (Автономный режим)
	Encryption     bool   // T1056 (Шифрование)
	PrinterAutomat bool   // T1221 (Принтер в автомате)

	// Вложенные теги (Nested Tags)
	OrgName     string // T1048
	Address     string // T1009
	Place       string // T1187
	OfdName     string // T1046
	OfdInn      string // T1017
	FnsSite     string // T1060
	SenderEmail string // T1117
}

// ItemPosition содержит параметры позиции чека.
type ItemPosition struct {
	Name     string  // Наименование товара
	Price    float64 // Цена
	Quantity float64 // Количество
	Tax      int     // Налоговая ставка
}

// PaymentInfo содержит параметры оплаты.
type PaymentInfo struct {
	Type int     // Тип оплаты (0 - наличные, 1 - безналичные, ...)
	Sum  float64 // Сумма
}

// DeviceOptions содержит настройки устройства (b0-b9).
type DeviceOptions struct {
	B0 int // Разделители
	B1 int // QR позиция
	B2 int // Округление
	B3 int // Авто-резак
	B4 int // Авто-тест
	B5 int // Открытие ящика (триггер)
	B6 int // Звук конца бумаги
	B7 int // Текст рядом с QR
	B8 int // Печать кол-ва покупок
	B9 int // Базовая СНО
}

// ReportKind определяет тип отчета.
type ReportKind string

const (
	ReportKindRegistration ReportKind = "registration"
	ReportKindCloseFn      ReportKind = "close_fn"
	// Другие типы отчетов можно добавить здесь
)

// Константы для типов отчетов
const (
	ReportReg   ReportKind = ReportKindRegistration
	ReportRereg ReportKind = ReportKindRegistration
)

// ReportMeta содержит метаданные и данные для отчета.
type ReportMeta struct {
	Kind  ReportKind
	Title string
	Data  interface{} // Данные отчета, например RegData или ReportFnCloseData
	Text  string      // Готовый текст отчета, если уже сформирован
}

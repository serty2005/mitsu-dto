package driver

// FiscalInfo содержит агрегированную информацию о фискальном регистраторе.
type FiscalInfo struct {
	ModelName        string `json:"modelName"`
	SerialNumber     string `json:"serialNumber"`
	RNM              string `json:"RNM"`
	OrganizationName string `json:"organizationName"`
	Address          string `json:"address"`
	Inn              string `json:"INN"`
	FnSerial         string `json:"fn_serial"`
	RegistrationDate string `json:"datetime_reg"`
	FnEndDate        string `json:"dateTime_end"`
	OfdName          string `json:"ofdName"`
	SoftwareDate     string `json:"bootVersion"`
	FfdVersion       string `json:"ffdVersion"`
	FnExecution      string `json:"fnExecution"`
	FnEdition        string `json:"fn_edition"`
	AttributeExcise  bool   `json:"attribute_excise"`
	AttributeMarked  bool   `json:"attribute_marked"`
}

// PrinterSettings содержит настройки принтера.
type PrinterSettings struct {
	Model    string `xml:"PRINTER,attr"`
	BaudRate int    `xml:"BAUDRATE,attr"`
	Paper    int    `xml:"PAPER,attr"`
	Font     int    `xml:"FONT,attr"`
	Width    int    `xml:"WIDTH,attr"`
	Length   int    `xml:"LENGTH,attr"`
}

// DrawerSettings содержит настройки денежного ящика.
type DrawerSettings struct {
	Pin  int `xml:"CD:PIN,attr"`
	Rise int `xml:"RISE,attr"`
	Fall int `xml:"FALL,attr"`
}

// ClicheLineData содержит данные одной строки клише.
type ClicheLineData struct {
	Text   string
	Format string // Строка вида "000000"
}

// LanSettings содержит настройки сети.
type LanSettings struct {
	Addr string `xml:"LAN,attr"`
	Port int    `xml:"PORT,attr"`
	Mask string `xml:"MASK,attr"`
	Dns  string `xml:"DNS,attr"`
	Gw   string `xml:"GW,attr"`
}

// OfdSettings содержит настройки ОФД.
type OfdSettings struct {
	Addr     string `xml:"OFD,attr"`
	Port     int    `xml:"PORT,attr"`
	Client   string `xml:"CLIENT,attr"`
	TimerFN  int    `xml:"TimerFN,attr"`
	TimerOFD int    `xml:"TimerOFD,attr"`
}

// ServerSettings содержит настройки сервера (OISM, OKP).
type ServerSettings struct {
	Addr string `xml:"ADDR,attr"` // Для OISM
	Okp  string `xml:"OKP,attr"`  // Для OKP (имя атрибута отличается)
	Port int    `xml:"PORT,attr"`
}

// TaxRates содержит настройки налоговых ставок.
type TaxRates struct {
	T1  string `xml:"T1,attr"`  // 20%
	T2  string `xml:"T2,attr"`  // 10%
	T3  string `xml:"T3,attr"`  // 20/120
	T4  string `xml:"T4,attr"`  // 10/110
	T5  string `xml:"T5,attr"`  // 0%
	T6  string `xml:"T6,attr"`  // Без НДС
	T7  string `xml:"T7,attr"`  // 5%
	T8  string `xml:"T8,attr"`  // 7%
	T9  string `xml:"T9,attr"`  // 5/105
	T10 string `xml:"T10,attr"` // 7/107
}

// RegData содержит полные данные о регистрации ККТ.
type RegData struct {
	RNM        string `xml:"T1037,attr"`
	Inn        string `xml:"T1018,attr"`
	FfdVer     string `xml:"T1209,attr"`
	RegDate    string `xml:"DATE,attr"`
	RegTime    string `xml:"TIME,attr"`
	RegNumber  string `xml:"REG,attr"`   // Порядковый номер регистрации
	Base       string `xml:"BASE,attr"`  // Коды причин регистрации
	TaxSystems string `xml:"T1062,attr"` // СНО (0,1,...)
	TaxBase    string `xml:"T1062_Base,attr"`

	// Маски режимов
	Mode    string `xml:"MODE,attr"`
	ExtMode string `xml:"ExtMODE,attr"`

	// Флаги (атрибуты)
	MarkAttr      string `xml:"MARK,attr"`  // Маркировка
	ExciseAttr    string `xml:"T1207,attr"` // Подакцизные
	InternetAttr  string `xml:"T1108,attr"` // Интернет
	ServiceAttr   string `xml:"T1109,attr"` // Услуги
	BsoAttr       string `xml:"T1110,attr"` // БСО
	LotteryAttr   string `xml:"T1126,attr"` // Лотерея
	GamblingAttr  string `xml:"T1193,attr"` // Азартные
	PawnAttr      string `xml:"PAWN,attr"`  // Ломбард
	InsAttr       string `xml:"INS,attr"`   // Страхование
	DineAttr      string `xml:"DINE,attr"`  // Общепит
	OptAttr       string `xml:"OPT,attr"`   // Опт
	VendAttr      string `xml:"VEND,attr"`  // Вендинг
	AutoModeAttr  string `xml:"T1001,attr"` // Автоматический режим
	AutoNumAttr   string `xml:"T1036,attr"` // Номер автомата (атрибут)
	AutonomAttr   string `xml:"T1002,attr"` // Автономный
	EncryptAttr   string `xml:"T1056,attr"` // Шифрование
	PrintAutoAttr string `xml:"T1221,attr"` // Принтер в автомате

	// Вложенные теги
	OrgName     string `xml:"T1048"`
	Address     string `xml:"T1009"`
	Place       string `xml:"T1187"`
	OfdName     string `xml:"T1046"`
	OfdInn      string `xml:"T1017,attr"`
	Site        string `xml:"T1060"`
	EmailSender string `xml:"T1117"`
	// T1036 может быть и тегом
	AutoNumTag string `xml:"T1036"`
}

// ShiftStatus содержит информацию о текущей смене.
type ShiftStatus struct {
	ShiftNum int    `xml:"SHIFT,attr"`
	State    string `xml:"STATE,attr"` // 0-закрыта, 1-открыта, 9-истекла
	Count    int    `xml:"COUNT,attr"`
	FdNum    int    `xml:"FD,attr"`
	KeyValid int    `xml:"KeyValid,attr"`

	// Вложенная структура для статуса обмена ОФД
	Ofd struct {
		Count int    `xml:"COUNT,attr"`
		First int    `xml:"FIRST,attr"`
		Date  string `xml:"DATE,attr"`
		Time  string `xml:"TIME,attr"`
	} `xml:"OFD"`
}

// ShiftTotals содержит итоги смены.
type ShiftTotals struct {
	ShiftNum int `xml:"SHIFT,attr"`
	Income   struct {
		Count string `xml:"COUNT,attr"`
		Total string `xml:"TOTAL,attr"`
	} `xml:"INCOME"`
	Payout struct {
		Count string `xml:"COUNT,attr"`
		Total string `xml:"TOTAL,attr"`
	} `xml:"PAYOUT"`
	Cash struct {
		Total string `xml:"TOTAL,attr"`
	} `xml:"CASH"`
}

// FnStatus содержит информацию о фискальном накопителе.
type FnStatus struct {
	Serial  string `xml:"FN,attr"`
	Ffd     string `xml:"FFD,attr"`
	Phase   string `xml:"PHASE,attr"`
	Valid   string `xml:"VALID,attr"`
	LastFD  int    `xml:"LAST,attr"`
	Flag    string `xml:"FLAG,attr"` // HEX маска предупреждений
	Edition string `xml:"EDITION,attr"`
}

// OfdExchangeStatus содержит информацию об обмене с ОФД.
type OfdExchangeStatus struct {
	Count    int    `xml:"COUNT,attr"`
	FirstDoc int    `xml:"FIRST,attr"`
	Date     string `xml:"DATE,attr"`
	Time     string `xml:"TIME,attr"`
}

// MarkingStatus содержит информацию о маркировке.
type MarkingStatus struct {
	MarkState int    `xml:"MARK,attr"`
	Keep      int    `xml:"KEEP,attr"`
	Flag      string `xml:"FLAG,attr"`
	Notice    int    `xml:"NOTICE,attr"`
	Holds     int    `xml:"HOLDS,attr"`
	Pending   int    `xml:"PENDING,attr"`
	Warning   int    `xml:"WARNING,attr"`
}

// RegistrationRequest содержит параметры для регистрации ККТ.
type RegistrationRequest struct {
	// Атрибуты (Attributes)
	IsReregistration bool   `json:"-"` // false = Регистрация, true = Перерегистрация
	Base             string `json:"-"` // Коды причин перерегистрации (через запятую, напр "1,5")

	RNM            string `json:"rnm"`             // T1037 (Рег. номер)
	Inn            string `json:"inn"`             // T1018 (ИНН Пользователя)
	FfdVer         string `json:"ffd_ver"`         // T1209 (Версия ФФД)
	TaxSystems     string `json:"tax_systems"`     // T1062 (СНО через запятую: "0,1")
	TaxSystemBase  string `json:"tax_base"`        // T1062_Base (Базовая СНО - опционально)
	AutomatNumber  string `json:"automat_num"`     // T1036 (Номер автомата)
	InternetCalc   bool   `json:"internet_calc"`   // T1108 (Расчеты в Интернет)
	Service        bool   `json:"service"`         // T1109 (Услуги)
	BSO            bool   `json:"bso"`             // T1110 (БСО)
	Lottery        bool   `json:"lottery"`         // T1126 (Лотерея)
	Gambling       bool   `json:"gambling"`        // T1193 (Азартные игры)
	Excise         bool   `json:"excise"`          // T1207 (Подакцизные товары)
	Marking        bool   `json:"marking"`         // MARK (Маркировка)
	PawnShop       bool   `json:"pawn_shop"`       // PAWN (Ломбард)
	Insurance      bool   `json:"insurance"`       // INS (Страхование)
	Catering       bool   `json:"catering"`        // DINE (Общепит)
	Wholesale      bool   `json:"wholesale"`       // OPT (Опт)
	Vending        bool   `json:"vending"`         // VEND (Вендинг)
	AutomatMode    bool   `json:"automat_mode"`    // T1001 (Автоматический режим)
	AutonomousMode bool   `json:"autonomous_mode"` // T1002 (Автономный режим)
	Encryption     bool   `json:"encryption"`      // T1056 (Шифрование)
	PrinterAutomat bool   `json:"printer_automat"` // T1221 (Принтер в автомате)

	// Вложенные теги (Nested Tags)
	OrgName     string `json:"org_name"`     // T1048
	Address     string `json:"address"`      // T1009
	Place       string `json:"place"`        // T1187
	OfdName     string `json:"ofd_name"`     // T1046
	OfdInn      string `json:"ofd_inn"`      // T1017
	FnsSite     string `json:"fns_site"`     // T1060
	SenderEmail string `json:"sender_email"` // T1117
}

// ItemPosition содержит параметры позиции чека.
type ItemPosition struct {
	Name     string  `json:"name"`     // Наименование товара
	Price    float64 `json:"price"`    // Цена
	Quantity float64 `json:"quantity"` // Количество
	Tax      int     `json:"tax"`      // Налоговая ставка
}

// PaymentInfo содержит параметры оплаты.
type PaymentInfo struct {
	Type int     `json:"type"` // Тип оплаты (0 - наличные, 1 - безналичные, ...)
	Sum  float64 `json:"sum"`  // Сумма
}

// DeviceOptions содержит настройки устройства (b0-b9).
type DeviceOptions struct {
	B0 int `xml:"b0,attr"` // Разделители
	B1 int `xml:"b1,attr"` // QR позиция
	B2 int `xml:"b2,attr"` // Округление
	B3 int `xml:"b3,attr"` // Авто-резак
	B4 int `xml:"b4,attr"` // Авто-тест
	B5 int `xml:"b5,attr"` // Открытие ящика (триггер)
	B6 int `xml:"b6,attr"` // Звук конца бумаги
	B7 int `xml:"b7,attr"` // Текст рядом с QR
	B8 int `xml:"b8,attr"` // Печать кол-ва покупок
	B9 int `xml:"b9,attr"` // Базовая СНО
}

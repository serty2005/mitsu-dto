package mitsu

// DevResponse содержит ответ на команду GET DEV (См. п. 3.3)
type DevResponse struct {
	Model string `xml:"DEV,attr"`
}

// VerResponse содержит ответ на команду GET VER (См. п. 3.4)
type VerResponse struct {
	Version string `xml:"VER,attr"`
	Size    string `xml:"SIZE,attr"`
	Crc32   string `xml:"CRC32,attr"`
	Serial  string `xml:"SERIAL,attr"`
	Mac     string `xml:"MAC,attr"`
	Status  string `xml:"STS,attr"`
}

// DateTimeResponse содержит ответ на команду GET DATE/TIME (См. п. 3.5)
type DateTimeResponse struct {
	Date string `xml:"DATE,attr"`
	Time string `xml:"TIME,attr"`
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

// OismSettings содержит настройки сервера OISM.
type OismSettings struct {
	Addr string `xml:"OISM,attr"`
	Port int    `xml:"PORT,attr"`
}

// ServerSettings содержит настройки сервера OKP.
type ServerSettings struct {
	Addr string `xml:"ADDR,attr"`
	Okp  string `xml:"OKP,attr"`
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

// RegData содержит полные данные о регистрации ККТ (См. п. 3.16).
type RegData struct {
	RNM        string `xml:"T1037,attr"`
	Inn        string `xml:"T1018,attr"`
	FfdVer     string `xml:"T1209,attr"`
	RegDate    string `xml:"DATE,attr"`
	RegTime    string `xml:"TIME,attr"`
	RegNumber  string `xml:"REG,attr"`   // Порядковый номер регистрации
	FdNumber   string `xml:"FD,attr"`    // Номер фискального документа
	FpNumber   string `xml:"T1077,attr"` // Фискальный признак из T1077
	Base       string `xml:"BASE,attr"`  // Коды причин регистрации
	TaxSystems string `xml:"T1062,attr"` // СНО (0,1,...)
	TaxBase    string `xml:"T1062_Base,attr"`

	ModeMask    uint32 `xml:"MODE,attr"`
	ExtModeMask uint32 `xml:"ExtMODE,attr"`

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

	OrgName     string `xml:"T1048"`
	Address     string `xml:"T1009"`
	Place       string `xml:"T1187"`
	OfdName     string `xml:"T1046"`
	OfdInn      string `xml:"T1017,attr"`
	Site        string `xml:"T1060"`
	EmailSender string `xml:"T1117"`
	AutoNumTag  string `xml:"T1036"`

	FnSerial      string // Заводской номер ФН
	FnEdition     string // Исполнение ФН
	PrinterSerial string // Серийный номер ФР
}

// ShiftStatus содержит информацию о текущей смене.
type ShiftStatus struct {
	ShiftNum int    `xml:"SHIFT,attr"`
	State    string `xml:"STATE,attr"` // 0-закрыта, 1-открыта, 9-истекла
	Count    int    `xml:"COUNT,attr"`
	FdNum    int    `xml:"FD,attr"`
	KeyValid int    `xml:"KeyValid,attr"`

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
	Power   string `xml:"POWER,attr"` // Флаг питания (1 = установлен, 0 = сброшен)
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

// RegResponse содержит ответ на команду регистрации/перерегистрации.
type RegResponse struct {
	FdNumber string `xml:"FD,attr"`    // Номер фискального документа
	FpNumber string `xml:"T1077,attr"` // Фискальный признак из T1077
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

// CashierResponse содержит ответ на команду GET CASHIER (См. п. 3.6)
type CashierResponse struct {
	Name string `xml:"CASHIER,attr"`
	Inn  string `xml:"INN,attr"`
}

// ComSettingsResponse содержит ответ на команду GET COM (См. п. 3.9)
type ComSettingsResponse struct {
	Speed int32 `xml:"COM,attr"`
}

// HeaderLine содержит данные одной строки заголовка/подвала
type HeaderLine struct {
	Text   string `xml:",chardata"`
	Format string `xml:"FORM,attr"` // Реальный вариант
	F      string `xml:"F,attr"`    // Старый вариант из документации
}

// HeaderResponse содержит ответ на команду GET HEADER (См. п. 3.10)
type HeaderResponse struct {
	L0 HeaderLine `xml:"L0"`
	L1 HeaderLine `xml:"L1"`
	L2 HeaderLine `xml:"L2"`
	L3 HeaderLine `xml:"L3"`
	L4 HeaderLine `xml:"L4"`
	L5 HeaderLine `xml:"L5"`
	L6 HeaderLine `xml:"L6"`
	L7 HeaderLine `xml:"L7"`
	L8 HeaderLine `xml:"L8"`
	L9 HeaderLine `xml:"L9"`
}

// PowerStatusResponse содержит ответ на команду GET POWER (См. п. 3.33)
type PowerStatusResponse struct {
	Power int `xml:"POWER,attr"`
}

// TimezoneResponse содержит ответ на команду GET TIMEZONE (См. п. 3.35)
type TimezoneResponse struct {
	Timezone string `xml:"TIMEZONE,attr"`
}

// CurrentDocumentTypeResponse содержит ответ на команду GET DOC='0'
type CurrentDocumentTypeResponse struct {
	Type int `xml:"TYPE,attr"`
}

// DocumentInfoResponse содержит информацию о документе (OFFSET и LENGTH) для команды GET DOC='X:{fd}'
type DocumentInfoResponse struct {
	Offset string `xml:"OFFSET,attr"`
	Length int    `xml:"LENGTH,attr"`
}

// ReadBlockResponse содержит ответ на команду READ (См. получения документа из ФН)
type ReadBlockResponse struct {
	Length int    `xml:"LENGTH,attr"`
	Data   string `xml:",innerxml"`
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

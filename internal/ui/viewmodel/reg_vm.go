package viewmodel

// RegistrationViewModel представляет модель данных для вкладки регистрации.
// Содержит все поля формы регистрации и информацию о ФН.
type RegistrationViewModel struct {
	// Основные данные регистрации
	RNM           string
	INN           string
	OrgName       string
	Address       string
	Place         string
	Email         string
	Site          string
	AutomatNumber string
	FFD           string
	Reasons       string

	// Оператор фискальных данных
	OFDName string
	OFDINN  string

	// Режимы работы (чекбоксы)
	ModeAutonomous bool
	ModeEncryption bool
	ModeService    bool
	ModeInternet   bool
	ModeAutomat    bool
	ModeBSO        bool
	ModeExcise     bool
	ModeGambling   bool
	ModeLottery    bool
	ModeMarking    bool
	ModePawn       bool
	ModeInsurance  bool
	ModeCatering   bool
	ModeWholesale  bool
	ModeVending    bool

	// Системы налогообложения
	TaxOSN        bool // 0 - Общая система налогообложения
	TaxUSN        bool // 1 - Упрощённая система налогообложения (доходы)
	TaxUSN_M      bool // 2 - Упрощённая система налогообложения (доходы минус расходы)
	TaxENVD       bool // 3 - Единый налог на вменённый доход
	TaxESHN       bool // 4 - Единый сельскохозяйственный налог
	TaxPat        bool // 5 - Патентная система налогообложения
	TaxSystemBase string

	// Информация о фискальном накопителе (ФН)
	FnNumber     string // Заводской номер ФН
	FnPhase      string // Фаза ФН (hex-строка)
	FnPhaseText  string // Текстовое описание фазы
	FnPhaseColor string // Цвет фазы (hex-строка, например "#FF0000")
	FnValidDate  string // Срок действия ФН
}

// NewRegistrationViewModel создаёт новый экземпляр RegistrationViewModel с дефолтными значениями.
func NewRegistrationViewModel() *RegistrationViewModel {
	return &RegistrationViewModel{
		FFD:           "1.05",
		FnPhaseText:   "—",
		FnValidDate:   "—",
		FnPhaseColor:  "#000000",
		TaxSystemBase: "",
	}
}

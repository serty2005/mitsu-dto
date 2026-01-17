package viewmodel

// MainViewModel отвечает за отображение статуса подключения и информации о ККТ на главной вкладке.
type MainViewModel struct {
	// Строка подключения (COMx:Baud или IP:Port)
	ConnectionString string

	// Список доступных подключений (Профили + COM порты) — ДОБАВЛЕНО
	ConnectionList []string

	// Статус подключения
	IsConnected bool

	// Текст кнопки действия (Подключить/Отключить/Искать)
	ActionButtonText string

	// Доступность элементов управления
	ConnectionStringEnabled    bool
	ActionButtonEnabled        bool
	ClearProfilesButtonEnabled bool

	// Информация о ККТ (показывается после подключения)
	KKTInfoVisible  bool
	ModelName       string
	SerialNumber    string
	UnsentDocsCount int
	PowerFlag       bool
}

// NewMainViewModel создаёт новый экземпляр MainViewModel с дефолтными значениями.
func NewMainViewModel() *MainViewModel {
	return &MainViewModel{
		ConnectionString:           "",
		ConnectionList:             []string{}, // Инициализация
		IsConnected:                false,
		ActionButtonText:           "Подключить",
		ConnectionStringEnabled:    true,
		ActionButtonEnabled:        true,
		ClearProfilesButtonEnabled: true,
		KKTInfoVisible:             false,
		ModelName:                  "Mitsu",
		SerialNumber:               "SN: ...",
		UnsentDocsCount:            0,
		PowerFlag:                  true,
	}
}

// UpdateUIState обновляет состояние интерфейса в зависимости от текущего статуса подключения.
func (vm *MainViewModel) UpdateUIState() {
	if vm.IsConnected {
		vm.ActionButtonText = "Отключить"
		vm.ConnectionStringEnabled = false
		vm.ActionButtonEnabled = true
		vm.ClearProfilesButtonEnabled = false
		vm.KKTInfoVisible = true
	} else {
		vm.ConnectionStringEnabled = true
		vm.ClearProfilesButtonEnabled = true

		text := trimSpace(vm.ConnectionString)
		if text == "" || text == "Поиск в сети / Ввести IP..." {
			vm.ActionButtonText = "Искать"
		} else {
			vm.ActionButtonText = "Подключить"
		}
		vm.ActionButtonEnabled = true
	}
}

func trimSpace(s string) string {
	result := []byte{}
	for _, c := range []byte(s) {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			result = append(result, c)
		}
	}
	return string(result)
}

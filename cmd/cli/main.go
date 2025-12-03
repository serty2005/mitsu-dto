package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"mitsuscanner/mitsu"
)

func main() {
	// КОНФИГУРАЦИЯ
	cfg := mitsu.Config{
		ConnectionType: 0,      // 0 - COM
		ComName:        "COM9", // <-- Убедитесь, что порт верный
		BaudRate:       115200,
		Timeout:        3000,
	}

	fmt.Printf("Подключение к %s...\n", cfg.ComName)
	drv := mitsu.New(cfg)

	if err := drv.Connect(); err != nil {
		log.Fatalf("Fatal: %v", err)
	}
	defer drv.Disconnect()
	fmt.Println("Подключено. Начинаем опрос всех методов...")

	// Вспомогательная функция для вывода
	printSection := func(name string, data interface{}, err error) {
		fmt.Printf("\n--- [%s] ---\n", name)
		if err != nil {
			fmt.Printf("ОШИБКА: %v\n", err)
			return
		}

		// Если это примитивный тип или строка
		switch v := data.(type) {
		case string, int, int32, time.Time, bool:
			fmt.Printf("Результат: %v\n", v)
		default:
			// Структуры выводим как JSON для читаемости
			b, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(b))
		}
	}

	// 1. Агрегированная инфо
	info, err := drv.GetFiscalInfo()
	printSection("GetFiscalInfo (Агрегированный)", info, err)

	// 2. Модель
	model, err := drv.GetModel()
	printSection("GetModel", model, err)

	// 3. Версия
	ver, serial, mac, err := drv.GetVersion()
	if err == nil {
		fmt.Printf("\n--- [GetVersion] ---\nВерсия: %s, Серийный номер: %s, MAC: %s\n", ver, serial, mac)
	} else {
		printSection("GetVersion", nil, err)
	}

	// 4. Дата и время
	dt, err := drv.GetDateTime()
	printSection("GetDateTime", dt, err)

	// 5. Кассир
	cashierName, cashierInn, err := drv.GetCashier()
	if err == nil {
		fmt.Printf("\n--- [GetCashier] ---\nКассир: %s, ИНН: %s\n", cashierName, cashierInn)
	} else {
		printSection("GetCashier", nil, err)
	}

	// 6. Настройки принтера
	printer, err := drv.GetPrinterSettings()
	printSection("GetPrinterSettings", printer, err)

	// 7. Денежный ящик
	drawer, err := drv.GetMoneyDrawerSettings()
	printSection("GetMoneyDrawerSettings", drawer, err)

	// 8. Скорость порта
	comSpeed, err := drv.GetComSettings()
	printSection("GetComSettings", comSpeed, err)

	// 9. Клише (запрашиваем 1-й вариант)
	header, err := drv.GetHeader(1)
	printSection("GetHeader (1)", header, err)

	// 10. Настройки LAN
	lan, err := drv.GetLanSettings()
	printSection("GetLanSettings", lan, err)

	// 11. Настройки ОФД
	ofd, err := drv.GetOfdSettings()
	printSection("GetOfdSettings", ofd, err)

	// 12. Настройки ОИСМ
	oism, err := drv.GetOismSettings()
	printSection("GetOismSettings", oism, err)

	// 13. Настройки ОКП
	okp, err := drv.GetOkpSettings()
	printSection("GetOkpSettings", okp, err)

	// 14. Налоги
	taxes, err := drv.GetTaxRates()
	printSection("GetTaxRates", taxes, err)

	// 15. Регистрационные данные
	reg, err := drv.GetRegistrationData()
	printSection("GetRegistrationData", reg, err)

	// 16. Статус смены
	shiftSt, err := drv.GetShiftStatus()
	printSection("GetShiftStatus", shiftSt, err)

	// 17. Итоги смены
	shiftTot, err := drv.GetShiftTotals()
	printSection("GetShiftTotals", shiftTot, err)

	// 18. Статус ФН
	fnSt, err := drv.GetFnStatus()
	printSection("GetFnStatus", fnSt, err)

	// 19. Обмен с ОФД
	ofdEx, err := drv.GetOfdExchangeStatus()
	printSection("GetOfdExchangeStatus", ofdEx, err)

	// 20. Маркировка
	markSt, err := drv.GetMarkingStatus()
	printSection("GetMarkingStatus", markSt, err)

	// 21. Часовая зона
	tz, err := drv.GetTimezone()
	printSection("GetTimezone", tz, err)

	fmt.Println("\n--- Установка кассира ---")
	if err := drv.SetCashier("Иванов И.И.", "344598649006"); err != nil {
		fmt.Printf("Ошибка установки кассира: %v\n", err)
	} else {
		fmt.Println("Кассир установлен успешно.")
	}

	// Синхронизация времени
	syncTimeUTC(drv)

	fmt.Println("\nТест завершен.")
}

func syncTimeUTC(drv mitsu.Driver) {
	fmt.Println("\n--- Синхронизация времени (UTC) ---")

	// Получаем текущее время в UTC
	now := time.Now().Local()
	fmt.Printf("Текущее время ПК (UTC): %s\n", now.Format("2006-01-02 15:04:05"))

	// Отправляем команду
	if err := drv.SetDateTime(now); err != nil {
		fmt.Printf("ОШИБКА установки времени: %v\n", err)
	} else {
		fmt.Println("Время успешно синхронизировано.")

		// Проверка
		checkTime, err := drv.GetDateTime()
		if err == nil {
			fmt.Printf("Время в кассе сейчас: %s\n", checkTime.Format("2006-01-02 15:04:05"))
		}
	}
}

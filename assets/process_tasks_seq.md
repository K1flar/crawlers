sequenceDiagram
    participant queue as Очередь сообщений <br/> Apache Kafka
    participant consumer as Consumer
    participant processor as Обработчик     
    participant db as База данных <br/> PostgreSQL
    participant crawler as Поисковый робот
    participant searx as Поисковая система <br/> SearXNG
    participant internet as Internet

    queue -->> consumer: Получение задачи из очереди
    consumer -->>+ processor: Process(id) <br/> запуск обработчика задачи
    
    processor ->> db: Получает задачу по ID 
    processor ->> db: Устанавливает статус задачи <br/> "В обработке"
    processor ->> db: Создает новый запуск для задачи <br/> фиксирует время начала
    processor ->>+ crawler: Start(task) <br/> запускает поискового робота
    
    crawler ->> searx: Получение стартовых URL по заданной теме
    loop
    crawler -->> internet: Установка соединения с очередным источником <br/> получение необходимой информации 
    internet -->> crawler: WEB-страница
    end 
    alt Обработка завершилась успешно
    crawler ->> processor: Список страниц со связями <br/> между источниками
    else Обработка завершилась с ошибкой
    crawler ->>- processor: Ошибка
    end

    alt Обработка завершилась успешно
    processor ->> db: Завершает запуск <br/> фиксирует время окончания обработки <br/> устанавливает статус "Завершен"
    processor ->> processor: Подсчитывает вес каждого доступного источника
    processor ->> db: Получает уже имеющиеся в системе <br/> источники по URL-адресу
    processor ->> processor: Фильтрация страниц <br/> определение какие создать, <br/> какие обновить
    processor ->> db: Создание новых источников
    processor ->> db: Обновление существующих источников
    processor ->> db: Создание связей между задачей <br/> и запуском с источниками
    processor ->> db: Устанавливает статус задачи <br/> "Активна"
    else Обработка завершилась с ошибкой
    processor ->> db: Завершает запуск <br/> фиксирует время окончания обработки <br/> устанавливает статус "Завершен с ошибкой"
    processor ->> db: Устанавливает статус задачи <br/> "Остановлена с ошибкой"
    end

    processor -->>- consumer: Завершение обработки
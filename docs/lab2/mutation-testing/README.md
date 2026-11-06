# ЛР2. Мутационное тестирование

Проверены `internal/domain/...` и `internal/usecase/...` на Gremlins v0.6.0
и Go 1.26.5. Исходный прогон выполнен на `d68a532`, повторный — на том же коде
с тремя новыми сценариями в `TestCreateWithKeyStoresNewSnapshot`.

## Отчёты до и после

| Gremlins | До | После |
| --- | ---: | ---: |
| Найдено позиций | 51 | 51 |
| `KILLED` | 42 | 43 |
| `LIVED` | 0 | 0 |
| `NOT COVERED` | 9 | 8 |
| Доля позиций, проверенных Gremlins | 82,35% | 84,31% |

В обоих прогонах было 0 таймаутов. Поле `mutants_total` в JSON считает
только мутантов, обработанных Gremlins. Полный список из 51 позиции лежит в `files`.
[До](before.json) · [После](after.json)

Gremlins использует покрытие операторов Go, чтобы отобрать мутации для
запуска. В профиле блоки `switch` начинаются после двоеточия `case`, а блоки
многострочных `if` могут заканчиваться до проверки возвращённой ошибки.
Поэтому статус `NOT COVERED` встречается даже у некоторых условий, которые
проверяют тесты. Инициализатор константы TTL также лежит вне блоков профиля.

Каждую из девяти исходных позиций `NOT COVERED` дополнительно проверил через
`go test -overlay` с одной подменой за запуск. Итог с учётом этой проверки
считает каждую позицию один раз.

| Gremlins и отдельные прогоны | До | После |
| --- | ---: | ---: |
| Тесты поймали мутацию | 45 | 49 |
| Тесты прошли с мутацией | 4 | 0 |
| Мутация не компилируется | 2 | 2 |

## Разбор позиций `NOT COVERED`

| Позиция | Подмена | До → после | Что проверяет тест |
| --- | --- | --- | --- |
| [Создание, 18:27](../../../internal/usecase/trip/create/usecase.go#L18) | `*` → `/` в `24 * time.Hour` | `LIVED` → `KILLED` | Снимок сохраняется с TTL ровно 24 часа. При делении TTL становится нулевым. |
| [Создание, 58:13](../../../internal/usecase/trip/create/usecase.go#L58) | `err == nil` → `err != nil` | `KILLED` → `KILLED` | Прежний тест требует вернуть сохранённый снимок для действующего ключа. |
| [Создание, 58:20](../../../internal/usecase/trip/create/usecase.go#L58) | `&&` → `\|\|` при проверке TTL | `LIVED` → `KILLED` | Просроченный ключ можно повторно использовать с другим телом. Мутация ошибочно возвращает конфликт. |
| [Создание, 64:13](../../../internal/usecase/trip/create/usecase.go#L64) | `err == nil` → `err != nil` в ветке удаления | `LIVED` → `KILLED` | Для нового ключа ожидается вставка снимка. Мутация вызывает удаление отсутствующей записи. |
| [Создание, 65:58](../../../internal/usecase/trip/create/usecase.go#L65) | `err != nil` → `err == nil` после удаления | `LIVED` → `KILLED` | Успешное удаление позволяет создать поездку, ошибка удаления прекращает операцию. Повторный Gremlins тоже поймал эту мутацию. |
| [Создание, 81:10](../../../internal/usecase/trip/create/usecase.go#L81) | `err != nil` → `err == nil` после транзакции | `KILLED` → `KILLED` | Прежний `TestCreate/success` требует успешного результата после коммита. |
| [Завершение, 40:10](../../../internal/usecase/trip/finish/usecase.go#L40) | `err != nil` → `err == nil` после транзакции | `KILLED` → `KILLED` | Прежний `TestFinish/success` требует вернуть завершённую поездку. |
| [Доменная ошибка, 11:17 и 11:24](../../../internal/domain/trip/errors.go#L11) | `+` → `-` при склейке строк | `NOT VIABLE` → `NOT VIABLE` | Компилятор отвергает вычитание строк в обеих позициях. |

## Воспроизведение

Из корня репозитория, на `d68a532` для отчёта «до» и с новым тестом для «после».
По `internal/usecase/trip/get` Gremlins не нашёл позиций для выбранных
операторов. Опция `--invert-logical` добавляет проверку `&&` и `||` к
стандартным операторам. [Документация Gremlins](https://gremlins.dev/latest/usage/commands/unleash/)

```sh
go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0
GOWORK=off GOMAXPROCS=2 gremlins unleash \
  --workers=4 --timeout-coefficient=15 --invert-logical \
  --exclude-files '(^|/)(cmd|docs)/|(^|/)internal/(config|container|generated|http|infra|integration|model|repository)/' \
  --output /tmp/lab2-gremlins.json --output-statuses lctv
```

Для отдельной проверки скопируйте исходный `.go` во временный файл и
подмените один оператор по таблице. В `overlay.json` задайте `Replace` с
абсолютным путём оригинала как ключом и путём временного файла как значением.
Запустите `go test -count=1 -failfast -json -overlay /path/to/overlay.json`
для пакета мутанта. В JSON падение теста содержит `Test`, а ошибка компиляции
появляется до запуска тестов.

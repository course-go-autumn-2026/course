// Слайд 19: LSP. Запуск из lecture: go run ./10-solid/04-lsp
// Вопрос: достаточно ли одинаковых сигнатур для корректной подстановки?
package main

import "fmt"

func main() {
	fmt.Println("До: проверяем общий контракт")
	BeforeLaunch(BeforeSparrow{})
	ShowBrokenSubstitution()

	fmt.Println("После: контракт отражает возможности")
	Launch(Sparrow{})
	WalkTogether(Sparrow{}, Penguin{})
	// Launch(Penguin{}) // Ошибка компиляции: у Penguin нет метода Fly.
	// Добавление пустого Fly лишь скрыло бы нарушение обещания «летать».
}

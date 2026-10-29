package warehouse

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type StorageCell interface {
	Name() string
	Price() int
}

type SmallCell struct{}
type LargeCell struct{}

type WithClimateControl struct{ Base StorageCell }
type WithInsurance struct{ Base StorageCell }

type HandlingPrice func(base int) int

var ErrUnknownStorageCell = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")

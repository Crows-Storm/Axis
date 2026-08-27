package menu

import "fmt"

type MenuType int8

const (
	MenuTypeDirectory MenuType = iota + 1
	MenuTypeMenu
	MenuTypeButton
)

func NewMenuType(t int8) (MenuType, error) {
	mt := MenuType(t)
	if err := mt.Validate(); err != nil {
		return 0, err
	}
	return mt, nil
}

func (m MenuType) Validate() error {
	switch m {
	case MenuTypeDirectory, MenuTypeMenu, MenuTypeButton:
		return nil
	default:
		return fmt.Errorf("invalid menu type: %d, must be 1(directory), 2(menu) or 3(button)", m)
	}
}

func (m MenuType) IsButton() bool {
	return m == MenuTypeButton
}

func (m MenuType) IsMenu() bool {
	return m == MenuTypeMenu
}

func (m MenuType) IsDirectory() bool {
	return m == MenuTypeDirectory
}

func (m MenuType) String() string {
	switch m {
	case MenuTypeDirectory:
		return "directory"
	case MenuTypeMenu:
		return "menu"
	case MenuTypeButton:
		return "button"
	default:
		return "unknown"
	}
}

package domain

// MenuSection represents a section of the menu.
type MenuSection struct {
	Category
	Products []Product
}

// Menu represents the whole menu.
type Menu []MenuSection

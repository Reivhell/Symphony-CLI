package ast

import "github.com/Reivhell/symphony/internal/blueprint"

// Injector adalah antarmuka untuk mesin pembedah dan penyisip kode.
type Injector interface {
	// Inject menyunting target source, melakukan inject terhadap anchornya.
	Inject(targetPath string, action blueprint.Action) error
	
	// CanHandle memberitahu engine jika engine injeksi ini mengenali tipe file tersebut
	CanHandle(filePath string) bool
}

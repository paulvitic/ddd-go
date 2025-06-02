package file

import "github.com/paulvitic/ddd-go"

type FilePersistenceConfig struct {
	DataDir string `json:"filePersitenceDir"`
}

func NewFilePersitenceConfig() *FilePersistenceConfig {
	return &FilePersistenceConfig{}
}

func (c *FilePersistenceConfig) OnInit() {
	config, err := ddd.Configuration[FilePersistenceConfig]("configs/properties.json")
	if err != nil {
		panic(err)
	}
	*c = *config
}

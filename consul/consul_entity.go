package consul

type ConsulEntity interface {
	GetName() string
}

func isWhitelisted(entity ConsulEntity) bool {
	whiteList := []string{"consul", "statsite", "operator"}
	for _, name := range whiteList {
		if entity.GetName() == name {
			return true
		}
	}

	return false
}

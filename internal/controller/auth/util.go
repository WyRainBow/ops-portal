package auth

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}


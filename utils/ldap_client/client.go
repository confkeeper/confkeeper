package ldap_client

import (
	"confkeeper/utils/config"
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"github.com/gookit/slog"
)

func LDAPAuth(username, password string) (bool, map[string]string, error) {
	var l *ldap.Conn
	var err error

	scheme := "ldap"
	if config.Cfg.Ldap.TLS {
		scheme = "ldaps"
	}
	ldapURL := fmt.Sprintf("%s://%s", scheme, config.Cfg.Ldap.Addr)

	var opts []ldap.DialOpt
	if config.Cfg.Ldap.TLS {
		opts = append(opts, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	l, err = ldap.DialURL(ldapURL, opts...)
	if err != nil {
		slog.Errorf("LDAP连接失败: %v", err)
		return false, nil, fmt.Errorf("LDAP连接失败")
	}
	defer l.Close()

	// Bind
	if config.Cfg.Ldap.BindDN != "" {
		err = l.Bind(config.Cfg.Ldap.BindDN, config.Cfg.Ldap.BindPass)
		slog.Errorf("LDAP绑定鉴权失败: %v", err)
		if err != nil {
			return false, nil, fmt.Errorf("LDAP绑定鉴权失败")
		}
	}

	// Search
	filter := fmt.Sprintf("(uid=%s)", username)
	searchReq := ldap.NewSearchRequest(
		config.Cfg.Ldap.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "cn", "mail", "uid"},
		nil,
	)

	sr, err := l.Search(searchReq)
	if err != nil {
		slog.Errorf("LDAP搜索失败: %v", err)
		return false, nil, fmt.Errorf("LDAP搜索失败")
	}

	if len(sr.Entries) != 1 {
		return false, nil, fmt.Errorf("未找到用户或找到多个用户")
	}

	userDN := sr.Entries[0].DN
	err = l.Bind(userDN, password)
	if err != nil {
		return false, nil, fmt.Errorf("LDAP认证失败")
	}

	userInfo := make(map[string]string)
	for _, attr := range sr.Entries[0].Attributes {
		if len(attr.Values) > 0 {
			userInfo[attr.Name] = attr.Values[0]
		}
	}

	return true, userInfo, nil
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
)

type LdapSearchOptions struct {
	host     string // ldap sever host:port
	username string // ldap username
	password string // ldap password

	baseDn string // base dn for search
}

type LdapSearchApp struct {
	opts *LdapSearchOptions
	cnn  *ldap.Conn
}

// 解析参数
func (lsa *LdapSearchApp) ParseOpts() (*LdapSearchOptions, error) {

	var opts LdapSearchOptions

	flag.Usage = func() {
		fmt.Printf("\nUsage: ldapsearch [-h host] [-u username] [-p password] [-b baseDn]\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&opts.host, "H", "ldap://192.168.1.1:389", "ldap server host format: ldap[s]://hostname:port/")
	flag.StringVar(&opts.username, "u", "", "username for authentication")
	flag.StringVar(&opts.password, "p", "", "password for authentication")
	flag.StringVar(&opts.baseDn, "b", "", "base DN for search")

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		return nil, errors.New(fmt.Sprintf("ParseOpts() error ; see usage for more information"))
	}

	return &opts, nil
}

// 连接 ldap 服务
func (lsa *LdapSearchApp) Connect() error {
	// ldap server host
	host := lsa.opts.host

	// 用来获取查询权限的 bind 用户.如果 ldap 禁止了匿名查询,那我们就需要先用这个帐户 bind 以下才能开始查询
	// bind 的账号通常要使用完整的 DN 信息.例如 cn=manager,dc=example,dc=org
	// 在 AD 上,则可以用诸如 mananger@example.org 的方式来 bind
	bindUsername := lsa.opts.username
	bindPassword := lsa.opts.password

	// 连接ldap
	cnn, err := ldap.DialURL(host)
	if err != nil {
		return err
	}

	err = cnn.Bind(bindUsername, bindPassword)
	if err != nil {
		return err
	}

	lsa.cnn = cnn

	log.Println("连接成功...")

	return err
}

// 搜索结构
func (lsa *LdapSearchApp) LdapSearch() {

	searchRequest := ldap.NewSearchRequest(
		//base dn,从此节点开始搜索
		lsa.opts.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		//查询返回的属性,以数组形式提供.如果为空则会返回所有的属性
		[]string{"dn", "cn", "objectClass"},
		nil,
	)

	searchResult, err := lsa.cnn.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	for _, item := range searchResult.Entries {
		item.PrettyPrint(4)
	}

	defer lsa.cnn.Close()
}

func main() {
	var (
		err error

		l LdapSearchApp
	)

	// 解析参数
	l.opts, err = l.ParseOpts()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 连接 ldap 服务
	err = l.Connect()
	if err != nil {
		fmt.Println("connect error: ", err)
		return
	}

	l.LdapSearch()
}

module github.com/goharbor/harbor/src

go 1.14

require (
	github.com/BurntSushi/toml v0.3.0 // indirect
	github.com/Masterminds/semver v1.4.2
	github.com/Sirupsen/logrus v1.0.5 //indirect
	github.com/Unknwon/goconfig v0.0.0-20160216183935-5f601ca6ef4d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/astaxie/beego v1.12.1
	github.com/beego/i18n v0.0.0-20140604031826-e87155e8f0c0
	github.com/beorn7/perks v1.0.0 // indirect
	github.com/casbin/casbin v1.8.3 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/dghubble/sling v1.1.0
	github.com/dgrijalva/jwt-go v3.0.0+incompatible
	github.com/docker/distribution v2.6.2+incompatible
	github.com/docker/go v1.5.1-1.0.20160303222718-d30aec9fd63c // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/docker/notary v0.5.1
	github.com/garyburd/redigo v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gocraft/work v0.5.1
	github.com/goharbor/harbor/tests/apitests/apilib v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate v3.3.0+incompatible
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.6.0
	github.com/lib/pq v1.0.0
	github.com/m3db/prometheus_client_golang v0.8.1 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/miekg/pkcs11 v0.0.0-20170220202408-7283ca79f35e // indirect
	github.com/opencontainers/go-digest v1.0.0-rc0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron v1.0.0
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/pflag v1.0.1 // indirect
	github.com/stretchr/testify v1.2.2
	github.com/theupdateframework/notary v0.5.1 // indirect
	github.com/uber-go/tally v3.3.11+incompatible
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550
	golang.org/x/oauth2 v0.0.0-20170928010508-bb50c06baba3
	google.golang.org/appengine v1.0.1-0.20170921170648-24e4144ec923 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20150924051756-4e86f4367175 // indirect
	gopkg.in/ldap.v2 v2.5.0
	gopkg.in/yaml.v2 v2.2.1
	k8s.io/apimachinery v0.0.0-20180704011316-f534d624797b // indirect
	k8s.io/client-go v8.0.0+incompatible // indirect
	k8s.io/helm v2.9.1+incompatible
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.3+incompatible
	github.com/goharbor/harbor => ../
	github.com/goharbor/harbor/tests/apitests/apilib => ../tests/apitests/apilib
	google.golang.org/api => google.golang.org/api v0.0.0-20160322025152-9bf6e6e569ff
)

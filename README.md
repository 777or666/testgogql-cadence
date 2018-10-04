# tetstgogql-cadence

Генерация моделей осуществляется приложением: gqlgen.exe
Конфигурация: gqlgen.yml

1 Переходим в папку
GOPATH\go\src\github.com\99designs\gqlgen

2 Запускаем gqlgen с параметром -с и указываем путь к файлу gqlgen.yml
gqlgen -c GOPATH\go\src\github.com\777or666\testgogql-cadence\models\gqlgen.yml


Ссылка на исходники: https://github.com/99designs/gqlgen
Сайт разработчиков: https://gqlgen.com/

Необходимо установить зависимости:
github.com/agnivade/levenshtein
github.com/go-resty/resty
github.com/gorilla/mux
github.com/hashicorp/golang-lru
github.com/rs/cors
sourcegraph.com/sourcegraph/appdash
github.com/google/protobuf
sourcegraph.com/sourcegraph/appdash-data



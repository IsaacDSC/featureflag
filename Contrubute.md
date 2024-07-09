# Contribute

## Tasks
    - Não manter estado - SERVICE
    - Poder utilizar mais de um adapter de armazenamento de banco de dados - SERVICE
    - Config authenticação e autorização - SERVICE
    - Implementar autenticação e autorização no SDK
    - Interface para listar/criar/editar/deletar FF no SERVICE

### mocks

``` sh
mockgen -source=./internal/domain/interfaces/featureflag_interfaces.go -destination=./internal/mocks/featureflag_interfaces.mock.go -package=mock
```
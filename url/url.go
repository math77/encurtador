package url

import (
  "math/rand"
  "net/url"
  "time"
)

//tamanho -> define o tamanho do id curto das urls
//simbolos -> define os simbolos que podem ser usados no id.
const (
  tamanho = 5
  simbolos = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_-+"
)

func init(){
  rand.Seed(time.Now().UnixNano())
}

type Url struct {
  Id string `json:"id"`
  Criacao time.Time `json:"criacao"`
  Destino string `json:"destino"`
}

type Stats struct {
  Url *Url `json:"url"`
  Clicks int `json:"clicks"`
}

//Especifica todas as operaçoes que um repositorio de urls é capaz de implementar.
type Repositorio interface {
  IdExiste(id string) bool
  BuscarPorId(id string) *Url
  BuscarPorUrl(url string) *Url
  Salvar(url Url) error
  RegistrarClick(id string)
  BuscarClicks(id string) int
}

var repo Repositorio

func ConfigurarRepositorio(r Repositorio) {
  repo = r
}

func (u *Url) Stats() *Stats {
  clicks := repo.BuscarClicks(u.Id)
  return &Stats{u, clicks}
}

func BuscarOuCriarNovaUrl(destino string) (u *Url, nova bool, err error) {
  if u = repo.BuscarPorUrl(destino); u != nil {
    return u, false, nil
  }

  if _, err = url.ParseRequestURI(destino); err != nil {
    return nil, false, err
  }

  url := Url{gerarId(), time.Now(), destino}
  repo.Salvar(url)
  return &url, true, nil
}

func gerarId() string {
  novoId := func() string {
    //como cada caracter é acessado individualmente eles são representados como bytes
    id := make([]byte, tamanho, tamanho)
    for i := range id {
      id[i] = simbolos[rand.Intn(len(simbolos))]
    }
    return string(id)
  }
  for {
    if id := novoId(); !repo.IdExiste(id) {
      return id
    }
  }
}

func Buscar(id string) *Url {
  return repo.BuscarPorId(id)
}

func RegistrarClick(id string) {
  repo.RegistrarClick(id)
}

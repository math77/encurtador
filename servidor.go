package main

import (
  "encoding/json"
  "fmt"
  "log"
  "flag"
  "net/http"
  "strings"

  "github.com/math77/encurtador/url"
)

var (
  porta *int
  logLigado *bool
  urlBase string
)

//Inicialização de var e recursos utilizados em um pacote.
func init() {
  porta = flag.Int("p", 8888, "porta")
  logLigado = flag.Bool("l", true, "log ligado/desligado")

  flag.Parse()

  urlBase = fmt.Sprintf("http://localhost:%d", *porta)
}

type Headers map[string]string

type Redirecionador struct {
  stats chan string
}

func (red *Redirecionador) ServeHTTP(w http.ResponseWriter, r *http.Request){

  buscarUrlEExecutar(w, r, func(url *url.Url){
    http.Redirect(w, r, url.Destino, http.StatusMovedPermanently)
    red.stats <- url.Id
  })
}

func Visualizador(w http.ResponseWriter, r *http.Request) {
  buscarUrlEExecutar(w, r, func(url *url.Url){
    json, err := json.Marshal(url.Stats())

    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    responderComJSON(w, string(json))
  })
}

func Encurtador(w http.ResponseWriter, r *http.Request) {

  //Verificando se a requisição enviada é do tipo POST.
  if r.Method != "POST" {
    responderCom(w, http.StatusMethodNotAllowed, Headers{
      "Allow": "POST",
    })
    return
  }

  url, nova, err := url.BuscarOuCriarNovaUrl(extrairUrl(r))

  if err != nil {
    responderCom(w, http.StatusBadRequest, nil)
    return
  }

  var status int
  if nova {
    status = http.StatusCreated
  } else {
    status = http.StatusOK
  }

  urlCurta := fmt.Sprintf("%s/r/%s", urlBase, url.Id)
  responderCom(w, status, Headers{
    "Location": urlCurta,
    "Link": fmt.Sprintf("<%s/api/stats/%s>; rel=\"stats\"", urlBase, url.Id),
  })
  gerarLogs("URL %s encurtada com sucesso para %s", url.Destino, urlCurta)
}

func buscarUrlEExecutar(w http.ResponseWriter, r *http.Request, executor func(*url.Url)) {
  caminho := strings.Split(r.URL.Path, "/")
  id := caminho[len(caminho)-1]

  if url := url.Buscar(id); url != nil {
    executor(url)
  } else {
    http.NotFound(w, r)
  }
}

func responderCom(w http.ResponseWriter, status int, headers Headers){
  //Itera sobre cada um dos cabeçalhos recebidos
  for k, v := range headers {
    //Configura cada um dos cabeçalhos
    w.Header().Set(k, v)
  }
  //Escreve o código de resposta final.
  w.WriteHeader(status)
}

func responderComJSON(w http.ResponseWriter, resposta string) {
  responderCom(w, http.StatusOK, Headers{
    "Content-Type": "application/json",
  })
  fmt.Fprintf(w, resposta)
}

func extrairUrl(r *http.Request) string {
  url := make([]byte, r.ContentLength, r.ContentLength)
  r.Body.Read(url)
  return string(url)
}

func registrarEstatisticas(ids <-chan string){
  for id := range ids {
    url.RegistrarClick(id)
    gerarLogs("Click registrado com sucesso para %s", id)
  }
}

func gerarLogs(formato string, valores ...interface{}){
  if *logLigado {
    log.Printf(fmt.Sprintf("%s\n", formato), valores...)
  }
}

func main() {
  url.ConfigurarRepositorio(url.NovoRepositorioMemoria())

  stats := make(chan string)
  defer close(stats)
  go registrarEstatisticas(stats)


  http.HandleFunc("/api/encurtar", Encurtador)
  http.HandleFunc("/api/stats/", Visualizador)
  http.Handle("/r/", &Redirecionador{stats})

  gerarLogs("Iniciando o servidor na porta %d...", *porta)
  log.Fatal(http.ListenAndServe(
    fmt.Sprintf(":%d", *porta), nil))
}

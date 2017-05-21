package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "encoding/base64"

)
import "io/ioutil"
import "strings"
import "strconv"

import "regexp"
type nSize struct{
   Alto  int     `json:"alto"`
   Ancho int     `json:"ancho"`
}
type bmpImg struct{
  Nombre string  `json:"nombre"`
  Data   string  `json:"data"`
  Size   nSize   `json:"tamaño"`
}
type Header struct { 
    tipo,reserved1, reserved2 uint16
    size, offset int    
}
type InfoHeader struct {         
    size, width,height,xresolution, yresolution, ncolours, importantcolours,compression, imagesize  int 
    planes, bits uint16
}
type RouteParams struct{
  Origen string  `json:"origen"`
  Destin string  `json:"destino"`
}
type RestaurantParams struct{
  Origen string  `json:"origen"`
}
type Coordinates struct{
  Lat   float64  `json:"lat"`
  Lng   float64  `json:"lng"`
}
type Routes struct{
  Locations []Coordinates `json:"ruta"`
}
type ErrorMsg struct{
  Status string           `json:"status"`
  Error  string           `json:"error"`
}

func Ejercicio1(key string, value interface{}, masterContainer, parentContainer, searchContainer, searchVariable string, w http.ResponseWriter) {

  var routes Routes
  for _, v := range value.([]interface{}) { 
      for k2, v2 := range v.(map[string]interface {}) {
          var coordinates Coordinates
          switch k2{
            case "end_location":
              coordinates.Lat = v2.(map[string]interface {})["lat"].(float64)
              coordinates.Lng = v2.(map[string]interface {})["lng"].(float64)
            case "start_location":
              coordinates.Lat = v2.(map[string]interface {})["lat"].(float64)
              coordinates.Lng = v2.(map[string]interface {})["lng"].(float64)
          }
          if(coordinates.Lat != 0.0 && coordinates.Lng != 0.0){
            routes.Locations = append(routes.Locations, coordinates) 
          }
      }
  }
  body, err := json.Marshal(routes)
  if err != nil {
      panic(err)
      WriteCurrentError("ERROR CREANDO RESPUESTA JSON", w)
  }
  w.Header().Set("Content-Type", "application/json")
  w.Write(body)
}
func Ejercicio2(key string, value interface{}, masterContainer, parentContainer, searchContainer, searchVariable string, w http.ResponseWriter) {
  var routes Routes
  for _, v := range value.([]interface{}) { 
    for k2, v2 := range v.(map[string]interface {}) {
        switch k2{
          case "geometry":
            for _, v3 := range v2.(map[string]interface {}) {
                var coordinates Coordinates
                if(v3.(map[string]interface {})["lng"] != nil && v3.(map[string]interface {})["lng"] != nil ){
                  coordinates.Lng = v3.(map[string]interface {})["lng"].(float64)
                  coordinates.Lat = v3.(map[string]interface {})["lat"].(float64)

                  if(coordinates.Lat != 0.0 && coordinates.Lng != 0.0){
                    routes.Locations = append(routes.Locations, coordinates) 
                  }
                }
            }
        }
        
    }
  }
  body, err := json.Marshal(routes)
  if err != nil {
      panic(err)
      WriteCurrentError("ERROR CREANDO RESPUESTA JSON", w)
  }
  w.Header().Set("Content-Type", "application/json")
  w.Write(body)
}
func WriteCurrentError(err string, w http.ResponseWriter){
  var errorMsg ErrorMsg 
  errorMsg.Status = "400"
  errorMsg.Error = err
  body, error := json.Marshal(errorMsg)
  if error != nil {
      panic(err)
  }
  w.WriteHeader(400)
  w.Write(body)

}
func decoding(baseImg string) []byte{
    data, _ := base64.StdEncoding.DecodeString(baseImg)
    return data
}
func getIValue(data []byte) (int) {
    var value int
    for _ ,element := range data {
        value |= int(element)
    }
    return value
}
func getIValue16(data []byte) (uint16) {
    var value uint16
    for _ ,element := range data {
        value |= uint16(element)
    }
    return value
}
func getHValues(header Header,  iheader InfoHeader, data []byte ) (Header, InfoHeader){
    header.offset=getIValue(data[10:14]) 
    iheader.width =getIValue(data[18:22]) 
    iheader.height=getIValue(data[22:26])
    iheader.bits=getIValue16(data[28:30])  
    return header, iheader
}

func readJson(mapArr map[string]interface{}, masterContainer, parentContainer, searchContainer, searchVariable string, w http.ResponseWriter )interface{}{
    var finalResult interface{}
    for key, value := range mapArr {
        if ( getVariable(parentContainer, key ,searchContainer,searchVariable) ){
            switch searchVariable{
              case "steps":
                Ejercicio1(key, value, masterContainer, parentContainer, searchContainer, searchVariable, w )
              case "results":
                Ejercicio2(key, value, masterContainer, parentContainer, searchContainer, searchVariable, w )
              case "location":
                fmt.Println(  value.(map[string]interface {})["lng"]  )
                fmt.Println(  value.(map[string]interface {})["lat"]  )
                Lng := ( value.(map[string]interface {})["lng"] ).(float64)
                Lat := ( value.(map[string]interface {})["lat"] ).(float64)
                strLng := strconv.FormatFloat(Lng, 'f', 6, 64)
                strLat := strconv.FormatFloat(Lat, 'f', 6, 64)
                fmt.Println( " CALL Ejercicio2 ", Lng, Lat, strLat, strLng )

                ConvertLoc("https://maps.googleapis.com/maps/api/place/nearbysearch/json?location="+strLat+","+strLng+"&radius=20000&type=restaurant&key=AIzaSyAUOMk8n8nhxUiSTEXq06jNth_kiV_s55E", w , "", "", "", "results")

            }
            return value
        }
        switch valueType := value.(type){
            case []interface{}:
                for key2, _ := range valueType{
                    switch valueType[key2].(type){
                        case string:
                        default:
                            masterContainer := parentContainer
                            parentContainer := key
                            readJson( valueType[key2].(map[string]interface{}) ,masterContainer, parentContainer, searchContainer,searchVariable , w)
                    }
                }
            case bool:
            case string:
            case float64:
            default:
                masterContainer := parentContainer
                parentContainer := key
                readJson( valueType.(map[string]interface{}) ,masterContainer, parentContainer, searchContainer,searchVariable , w)
            
        }
    }

    return finalResult
    
}
func getVariable(parentContainer, key ,searchContainer,searchVariable string) bool{
    if( (parentContainer == searchContainer) && (key == searchVariable) ){
        return true;
    }
    return false;
}

func ConvertLoc( url string, w http.ResponseWriter , masterContainer, parentContainer, searchContainer, searchVariable string ){
    resp, err := http.Get(url)
  if err != nil {
    WriteCurrentError("ERROR OBTENEIENDO LOCACION DEL URL", w)
    panic(err)

  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
    WriteCurrentError("ERROR CONVIRTIENDO EL JSON A BYTES", w)
    panic(err)
  }
   
    var arbitrary_json map[string]interface{}
    json.Unmarshal(body , &arbitrary_json)
  
    fmt.Println( "RESULTADO ")
    
    fmt.Println( readJson(arbitrary_json, masterContainer, parentContainer, searchContainer, searchVariable, w)  )
    
}


func handler(w http.ResponseWriter, r *http.Request) {
    
    defer r.Body.Close()
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        WriteCurrentError("ERROR CONVIRTIENDO EL JSON A BYTES", w)
        fmt.Println("CANT READ IT",err)
    }

    switch r.URL.Path[1:] {
      case "ejercicio1":
        var routes RouteParams
        err := json.Unmarshal(body, &routes)
        if err != nil {
            fmt.Println("CANT CCONVERT IT:", routes)
            WriteCurrentError("FORMATO JSON INVALIDO", w)
        }

        match1, _ := regexp.MatchString("([[:alnum:]]|,|[[:blank:]])+", routes.Origen )
        match2, _ := regexp.MatchString("([[:alnum:]]|,|[[:blank:]])+", routes.Destin )
        if(!match1 || !match2 ){
          WriteCurrentError("Direccion de origen o destino invalida", w)
          return
        }

        arr1 := strings.Split(routes.Origen, " ")
        arr2 := strings.Split(routes.Destin, " ")
        var address1 string = ""
        var address2 string = ""
        for index,element := range arr1 {
            address1 += element
            if(index < len(arr1)-1 ){
              address1 += "+"
            }
        }
        for index2,element2 := range arr2 {
            address2 += element2  
            if(index2 < len(arr2)-1 ){
              address2 += "+"
            }
        }
        ConvertLoc( "https://maps.googleapis.com/maps/api/directions/json?origin="+address1+"&destination="+address2+"&key=AIzaSyAUOMk8n8nhxUiSTEXq06jNth_kiV_s55E" , w , "", "", "legs", "steps" )
      
      case "ejercicio2":
        var restaurants RouteParams
        err := json.Unmarshal(body, &restaurants)
        if err != nil {
            WriteCurrentError("ERROR CONVIRTIENDO EL JSON A BYTES", w)
            fmt.Println("CANT CCONVERT IT:", restaurants)
        }
     
        arr1 := strings.Split(restaurants.Origen, " ")
        var address1 string = ""
        for index,element := range arr1 {
            address1 += element
            if(index < len(arr1)-1 ){
              address1 += "+"
            }
        }
        ConvertLoc( "https://maps.googleapis.com/maps/api/geocode/json?address="+address1+"&key=AIzaSyAUOMk8n8nhxUiSTEXq06jNth_kiV_s55E" , w , "", "", "geometry", "location" )

      
      case "ejercicio3":
        var img bmpImg
        err := json.Unmarshal(body, &img)
        if err != nil {
            panic(err)
        }
        data := decoding(img.Data) 

        var h Header
        var ih InfoHeader
        h, ih =  getHValues(h, ih, data)
        var headerOffset = 14 + 40 + 4*ih.ncolours
        fmt.Println(h,ih,  ih.bits,headerOffset )
        var tmpData []byte
        for i := 0; ((i + int(ih.bits) < len(data)) ); i++ {
            if( i >= h.offset ){
                var prom int
                for j := 0; j < int(ih.bits) ; j++ {
                    prom += int(data[i+j])
                }
                prom += prom/ int(ih.bits)
                for j := 0; j < int(ih.bits) ; j++ {
                    tmpData  = append(tmpData, byte(prom) )
                }
                i += int(ih.bits) - 1
            }else{
                tmpData  = append(tmpData , data[i])
            }
        }
  
        arr1 := strings.Split(img.Nombre, ".")
        img.Nombre = arr1[0]+"(Blanco -- Negro)."+arr1[1]
        uEnc := base64.URLEncoding.EncodeToString(tmpData)
        img.Data = uEnc
        sendBody, errr := json.Marshal(img)
        if errr != nil {
            panic(errr)
        }
        w.Header().Set("Content-Type", "application/json")
        w.Write(sendBody)
        fileErr2 := ioutil.WriteFile(img.Nombre, tmpData, 0644)
        if(fileErr2 != nil ){
            panic(fileErr2)
        }
    

      case "ejercicio4":
        var img bmpImg
        err2 := json.Unmarshal(body, &img)
        if err2 != nil {
            WriteCurrentError("ERROR CONVIRTIENDO EL JSON A BYTES", w)
            fmt.Println("CANT CCONVERT IT:", err2)
        }
        fmt.Println("BODY\n")
        fmt.Println("", []byte(img.Data))

        arr := strings.Split(img.Nombre, ".")
        img.Nombre = arr[0]+"(Reducida)."+arr[1]

        body, err := json.Marshal(img)
        if err != nil {
            WriteCurrentError("ERROR CONVIRTIENDO RESPUESTA A JSON", w)
            panic(err)
        }
     
        w.Header().Set("Content-Type", "application/json")
 
        w.Write(body)


    }

}

func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/ejercicio1", handler)
    http.HandleFunc("/ejercicio2", handler)
    http.HandleFunc("/ejercicio3", handler)
    http.HandleFunc("/ejercicio4", handler)
    http.ListenAndServe(":8080", nil)
}
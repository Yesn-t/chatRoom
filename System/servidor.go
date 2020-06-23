package main

import (
	"os"
	"fmt"
	"net"
	"bufio"
	"encoding/gob"
)

//----------------------------------- Global
const BUFFERSIZE = 1024
var UserMensaje Usuario
var ClientList []Client
var historial []string
var restantes int
var NextId int

//----------------------------------- Estructura
type Usuario struct {
	Id   int
	Name string
	Msg  string
	Bits []uint8
	Puerto string
	Activo bool
}

type Client struct {
	User Usuario
	Conn net.Conn
}

//----------------------------------- Servidor
func servidor(){
    s, err := net.Listen("tcp", ":9999")
    if err != nil {
        fmt.Println(err)
        return
    }
    for {
		c, err := s.Accept() // Espera a peticion
        if err != nil {
            fmt.Println(err)
            continue
		}
		go func (){
			for i := 0; i < 2; i++ {
				handleClient(c, UserMensaje)
			}
		}()
    }
}

func handleClient(c net.Conn, UserTemp Usuario){
    var usr Usuario
	tipo, err := bufio.NewReader(c).ReadString('\n')
	
	if err != nil {
			delete(UserTemp.Id) // Eliminar si no esta hay respuesta de conexion
			return
	}

	if tipo == "id\n" {				// Pedir Id de cliente
		gob.NewEncoder(c).Encode(NextId)
		NextId += 1
				
	} else if tipo == "abrir\n" { 	// Agregar cliente
		err = gob.NewDecoder(c).Decode(&usr)
				
		if err != nil {
			fmt.Println("Struct Abrir: ",err)
			return
		}

		fmt.Println("\n\n Conectado")
		fmt.Println(" Nick: " + usr.Name)

		createforlder(usr.Name)

		ClientList = append(ClientList, Client{User: usr, Conn: c})    
	
	} else if tipo == "texto\n"{  	// Recibir texto de cliente
		err = gob.NewDecoder(c).Decode(&usr)
		UserMensaje = usr
		mensaje := " " + UserMensaje.Name + ":  " + UserMensaje.Msg
		fmt.Print("\n " + mensaje)
		historial = append(historial, mensaje)

		if err != nil {
			fmt.Println("Texto: ",err)
			return
		}

		restantes = len(ClientList) - 1
		changeActive(UserTemp.Id)
		go propagate()

	} else if tipo == "archivo\n"{  // Recibir archivo de cliente

		err = gob.NewDecoder(c).Decode(&usr)
		UserMensaje = usr
		mensaje := " " + UserMensaje.Name + ":  " + UserMensaje.Msg
		fmt.Print("\n " + mensaje)
		historial = append(historial, mensaje)

		if err != nil {
			fmt.Println("Archivo: ",err)
			return
		}

		restantes = len(ClientList) - 1
		changeActive(UserTemp.Id)
		go propagate()

	} else if tipo == "cerrar\n"{
		c.Close()
		delete(UserTemp.Id)
		return
	} else {
		fmt.Println("Sin tipo de envio")  // Mensaje de no conexion
	}
}

func delete(id int){
    for i, x := range ClientList {
		if x.User.Id == id{
			fmt.Println("\n Desconectado: ", x.User.Name)
			for j := i; j < len(ClientList)-1; j++ {
				ClientList[j] = ClientList[j+1]
			}
			ClientList = ClientList[:len(ClientList)-1]
			break
		}
	}
}

//----------------------------------- Clientes
func changeActive(id int) {
	for i, x := range ClientList {
		if x.User.Id == id{
			ClientList[i].User.Activo = false
			break
		}
	}
}

func createforlder(userName string){
	path := "../Clients/" + userName
	_, err := os.Stat(path)
 
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			fmt.Println("\n Folder: ",err)
		}
	}
}

//----------------------------------- Chat
func detected() {
	for{
		for  i:=0 ; i<len(ClientList); i++ {
			if ClientList[i].User.Activo == false {
				ClientList[i].User.Activo = true
				go handleClient(ClientList[i].Conn, ClientList[i].User)
			}
		}
	}	
}

func send(indiceClient int) {
	p := ClientList[indiceClient].User.Puerto
	
	c, err := net.Dial("tcp", (":" + p))
	if err != nil {
		delete(ClientList[indiceClient].User.Id)
		fmt.Println("Conexion a Puerto: ", p, " U:", indiceClient, " -> ", err)
		return
	}

	usr := UserMensaje
	err = gob.NewEncoder(c).Encode(usr) 
	if err != nil {
		fmt.Println("Envio a Puerto: ", p, " U:", indiceClient, " -> ", err)
		return
	}
	c.Close()
}

func propagate(){
	for  i:=0 ; i<len(ClientList); i++ {
		if restantes > 0 && UserMensaje.Id != ClientList[i].User.Id {
			restantes -= 1
			go send(i)
		}
	}
}

//----------------------------------- Respaldo
func respaldo() {
	os.Remove("respaldo.txt")

	file, err := os.Create("respaldo.txt")
	if err != nil {
		fmt.Println("Respaldo:", err)
		return
	}

	for _, text := range historial {
		file.WriteString(text + " \n")
	}

	file.Close()
	fmt.Println("\n Repaldo Creado")
}

//----------------------------------- Main
func main() {
	NextId = 0
	restantes = 0
    salir := false
    var opcion int
    
	go servidor()
	go detected()

	for salir != true {
		fmt.Println("\n\n\n   Menu")
		fmt.Println(" 1  - Mostrar historial")
		fmt.Println(" 2  - Guardar historial")
		fmt.Println(" 3  - Salir")
		fmt.Print(" Opcion: ")
		fmt.Scan(&opcion)


		if opcion == 1 {
			fmt.Println("")
			fmt.Println(" Total clientes: ",len(ClientList))
			fmt.Println("\n\n Historial")
			for _, text := range historial{
				fmt.Println(" " + text)	
			} 
			fmt.Println("\n\n")
		
		} else if opcion == 2 {
			respaldo()
			
		} else if opcion == 3 {
			salir = true
			return

		} else {
			fmt.Println("Opcion incorrecta")
		}
	}
    
    var input string
    fmt.Scanln(&input)
}   
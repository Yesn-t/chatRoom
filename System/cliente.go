package main

import (
	"encoding/gob"
	"io/ioutil"
	"strconv"
	"bufio"
	"time"
	"fmt"
	"net"
	"os"
)

//----------------------------------- Global
const BUFFERSIZE = 1024
var historial []string
var nick string

//----------------------------------- Estructura
type Usuario struct {
	Id   int
	Name string
	Msg  string
	Bits []uint8
	Puerto string
	Activo bool
}

//----------------------------------- Cliente
func getID(c net.Conn) int{
	var id int
	fmt.Fprintf(c, "id\n")					// Tipo de conexion
    err := gob.NewDecoder(c).Decode(&id) 	// Datos

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return id
}

func cliente(c net.Conn, u Usuario){
	fmt.Fprintf(c, "abrir\n")			// Tipo de conexion
    err := gob.NewEncoder(c).Encode(u) 	// Datos
    
	if err != nil {
		fmt.Println(err)
		return
	}
}

func endConn(c net.Conn){
	fmt.Println(" Wait to close...")
	time.Sleep(time.Millisecond*1000)

	fmt.Fprintf(c, "cerrar\n")				// Tipo de conexion
    err := gob.NewEncoder(c).Encode("") 	// Datos
    
	if err != nil {
		fmt.Println(err)
		return
	}
	c.Close()
}

//----------------------------------- Chat
func listening(puerto string){
    s, err := net.Listen("tcp", (":" + puerto))
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
		go read(c)
    }
}

func read(c net.Conn) {
	var usr Usuario
	err := gob.NewDecoder(c).Decode(&usr)

	if err != nil {
		fmt.Println("Lectura de mensaje: ",err)
		return
	}

	if len(usr.Bits) > 0 {
		saveFile(usr.Msg, usr.Bits)
	}

	mensaje := " " + usr.Name + ":  " + usr.Msg
	historial = append(historial, mensaje)
	fmt.Print("\n" + mensaje)
}

func write(c net.Conn, tipo int, u Usuario, file string){
	if tipo == 1 {
		fmt.Fprintf(c, "texto\n")			// Envio de Texto
		u.Bits = nil
		err := gob.NewEncoder(c).Encode(u) 
		if err != nil {
			fmt.Println(" Escritura Msg:", err)
			return
		}	

	}else{
		fmt.Fprintf(c, "archivo\n")			// Envio de Archivo			
		u.Msg = file
		u.Bits = readFile(file)
		err := gob.NewEncoder(c).Encode(u)
		if err != nil {
			fmt.Println(" Escritura Arch:", err)
		}
	}
}

//----------------------------------- Archivos
func readFile(fileName string) []uint8{
	bs, err := ioutil.ReadFile("../Files/" + fileName)
	
	if err != nil {
		fmt.Println("No se pudo abrir el archivo")
		return nil
	}

	return bs
}

func saveFile(fileName string, data []uint8){
	save, err := os.Create("../Clients/" + nick + "/" + fileName)
	if err != nil {
		fmt.Println("No se pudo guardar")
		return
	}
	defer save.Close()

    save.Write(data)
}

//----------------------------------- Main
func menu(){
	fmt.Println("\n\n\n   Menu")
	fmt.Println(" 1  - Enviar Mensaje")
	fmt.Println(" 2  - Enviar Archivo")
	fmt.Println(" 3  - Mostrar Historial")
	fmt.Println(" 4  - Salir")
	fmt.Print(" Opcion: ")
}

func archivos(){
	fmt.Println("\n\n\n   Menu")
	fmt.Println(" 1  - Enviar Archivo Mat")
	fmt.Println(" 2  - Enviar Txt")
	fmt.Println(" 3  - Enviar Excel")
	fmt.Println(" 4  - Enviar PDF")
	fmt.Println(" 5  - Enviar IMG")
	fmt.Print(" Opcion: ")
}

func main() {
	salir := false
	var input string
	var opcion int
	c, err := net.Dial("tcp", ":9999")
	
	if err != nil {
        fmt.Println(" Conexion: ", err)
        return
	}
	defer endConn(c)

	fmt.Print(" Nick: ")
    fmt.Scanln(&nick)

	id := getID(c)
	puerto := strconv.Itoa( (9000 + id) )
	u := &Usuario{
		Id:		id, 
		Name:	nick,
		Msg:    "",
		Bits:   nil,
		Puerto: puerto,
		Activo: false,
	}

	go cliente(c, *u)
	go listening(puerto)

	for salir != true {
		menu()
		fmt.Scan(&opcion)


		if opcion == 1 {
			fmt.Scanln(&input)
			fmt.Print(" Escribe el mensaje: ")
			
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan() 
			line := scanner.Text()
			u.Msg = line
			
			write(c, 1, *u, "") 	

		} else if opcion == 2 {
			var archivo string

			archivos()
			fmt.Scanln(&input)
			fmt.Scan(&opcion)

			switch opcion {
                case 1 :    
                    archivo = "mat.mat"   
				case 2:
                    archivo = "texto.txt"
				case 3:
					archivo = "excel.xls"
				case 4:
					archivo = "pdf.pdf"
				case 5:
					archivo = "img.jpg"
            }
			
			write(c, 2, *u, archivo) 
			
		} else if opcion == 3 {
			fmt.Println("\n\n Historial")
			for _, text := range historial{
				fmt.Println(" " + text)	
			} 
			
		} else if opcion == 4 {
			salir = true
			return

		} else {
			fmt.Println("Opcion incorrecta")
		}
	}

	fmt.Scanln(&input)
}   
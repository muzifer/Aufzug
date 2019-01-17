package main

import (
	"fmt"
	"math/rand"
	"sync"
	
)

//-------------------------------------------STRUCTS-------------------------------------------
var wg sync.WaitGroup
var wg2 sync.WaitGroup

type Aufzug struct {
	aufzugNr int // aufzug Nr 
	fahrgaeste int	//Anzahl Passagiere
	max int  		//Maximalkapazität
	strecke int 	//Gesamt-Wegstrecke
	etage int // aktuelle Etage
	zielEtage int // nächster halt 
}
type Person struct {
	wartezeit int//Wartezeit in schritten
	soll int//Distanz start- ziel etage
	ist int// tatstächlich gefahrene Strecke
	abweichung int // differenz zwischen ist und soll
	start int // Startegage
	ziel int //Zieletage
	status int //  0 = wartend auf zuteilung, 1 = zugeteilt, 2 = fahrend, 3 = ausgestiegen
	aufzugNr int // nr des befördernden aufzugs
}
type Zentrale struct {
	anzSim int // anzahl an Simulationsläufen
	anzAufz int // anzahl Aufzüge
	maxPers int // maximale anzahl an personen pro simulation
	dauer int // dauer des simulationslaufs in Gesamtschritten 
	algNr int // aufzugalgorithmen (Nr)
	auswertung []Auswertung // auswertung weiteres struct ???
}
type Auswertung struct { // alle aufzüge addiert!!!
	wartezeitP int //gesamtwartezeit der passagiere in schritten
	abweichungP int // gesamtdifferenz von soll- zu ist-strecken der passagiere in schritten
	streckeA int // gesamtstrecke aller Aufzüge addiert
}




// ---------------------------------------------ZENTRALE STEUERLOGIK----------------------------------------
func ZentraleSteuerlogik() {

// woher bekommt die zentrale steuerlogik die aufzugalgorithmen?

	//slice erzeugen für auswertungen
	
	auswertungsListe := make([]Auswertung,0)
	
	anzahlSimulationsläufe := 3
	anzahlAufzüge := 3
	dauer := 10
	maxPersonen := 10
	
	// angegebene anzahl von Simulationsläufen aufrufen, auswertung wird als rückgabewert erhalten und an liste angehängt
	for i := 0; anzahlSimulationsläufe > i; i++{
		wg.Add(1)//DEADLOCK
		auswertung := Steuersimulation(anzahlAufzüge,dauer,maxPersonen)
		auswertungsListe = append(auswertungsListe,auswertung)
		fmt.Println("Auswertung:", auswertungsListe)
	}
}




//----------------------------------------------STEUERLOGIK EINES SIMULATIONSLAUFS---------------------------------------------------

func Steuersimulation (anz, dauer, maxPers int/*, algorithmus func()*/)(auswertung Auswertung){

	// slices erzeugen 
	aufzugListe := make([]Aufzug,0)
	fahrgaesteListe := make([]Person,0) 

	// channels erzeugen
	channelGP := make (chan Person)
	channelGA := make (chan Aufzug)
	channelAlgP := make (chan []Person)
	channelAlgA := make (chan []Aufzug)
	
	//generiere Aufzüge
	//wg.Add(1)
	go GeneriereAufzuege(anz, channelGA) // aufzüge erzeugen
	
	//empfange Nachrichten von generiere Aufzug, solange welche gesendet werden
	fmt.Println("steuersimulation vor empfange nachrichten generiege Aufzüge")
	
	for range channelGA{
		neuerAufzug := <-channelGA
		fmt.Println("neuer Aufzug: ", neuerAufzug)
		aufzugListe = append(aufzugListe, neuerAufzug)	
	}
	fmt.Println("steuersimulation NACH empfange nachrichten generiege aufzüge")
	fmt.Println("aufzugliste:", aufzugListe)

	//starte Simulation
	for ; dauer >= 0; dauer--{ // überwacht gesamtlänge der simulation
		
		fmt.Println("steuersimulation:for dauer")

		//erzeuge neue Personen
		if  maxPers > len(fahrgaesteListe){ //überprüfe ob maximale personenzahl erreicht ist, wenn noch nicht dann...
			fmt.Println("steuersimulation: if case Personenanzahl")
			//wg.Add(1)
			//generiere Passagiere
			go GenerierePassagiere(maxPers - len(fahrgaesteListe), channelGP)// übergibt berechnete maximal erlaubte personenanzahl
			
			//empfange Nachrichten von GenerierePassagiere
			fmt.Println("1000 steuersimulation vor empfange nachrichten generiege Passagiere")
			for range channelGP{
			neueAnfragen := <-channelGP // speichere Nachricht in variabel
			fahrgaesteListe = append(fahrgaesteListe, neueAnfragen)// hänge neue anfrage an fahrgästeliste an
			fmt.Println("fahrgästeliste: ", fahrgaesteListe)
			}	
			
			fmt.Println("steuersimulation NACH empfange nachrichten generiege Passagiere")
			
		
		}	

		wg2.Add(1)
		fmt.Println("VOR Algorithmus")
		// algorithmus aufrufen
		go Aufzugsteuerungs_Agorithmus_1(channelAlgA, channelAlgP, maxPers, dauer)			

		fmt.Println("NACH Algorithmus")
		// auswertung senden
		select{
		case aufzugListe = <- channelAlgA:aufzugListe = <- channelAlgA
		case fahrgaesteListe = <- channelAlgP: fahrgaesteListe = <- channelAlgP
		default:
		}
		
		
	}
	
	//nach beendigung der simulation
	// erstelle auswertung	
		for i := range aufzugListe{
			auswertung.streckeA += aufzugListe[i].strecke // gesamtstrecke aller Aufzüge addiert
		}
		for i := range fahrgaesteListe{
			auswertung.wartezeitP += fahrgaesteListe[i].wartezeit //gesamtwartezeit der passagiere in schritten
			auswertung.abweichungP += fahrgaesteListe[i].abweichung // gesamtdifferenz von soll- zu ist-strecken der passagiere in schritten
		}

	fmt.Println("HIIIIIIIIIEEEEEEEEEEERRRRRR",fahrgaesteListe)
	//wg.Wait()
	
	wg2.Wait()
	wg.Done() // DEADLOCK
	return // gibt auswertung des simulationslaufs zurück
	
}

//--------------------------------------------GENERIERE AUFZÜGE---------------------------------------------------
func GeneriereAufzuege (anzA int, channelGA chan Aufzug){

	maxPers := 1 // wieviele dürfen höchstens mitfahren
	fmt.Println("for: anz Aufzüge = äußere schleife:", anzA)
	for i := 1; anzA >= i; i++{
		
			
			fmt.Println("AufzugNr:",i)
			
			neuerAufzug := Aufzug{aufzugNr: i, max: maxPers} 
			fmt.Println("neuer Aufzug",neuerAufzug)
			
			channelGA <- neuerAufzug
	}
	
	//close(channelGA)
	//wg.Done()
}




//------------------------------------------GENERIERE PASSAGIERE-----------------------------------------------------

func GenerierePassagiere(max int, channelGP chan Person){ // hier kann anzahl an personen je schritt und anzahl etagen geändert werden
	fmt.Println("Einstieg generiere Passagiere")
	//wg.Add(1)
	maxAnz := 3
	
	if max < 3{ //maximal 3 personen je schritt
		maxAnz = max// zufällige anzahl an passagieren mit zufälligen start und ziel etagen erstellen
		
	}
	fmt.Println(maxAnz)
	anz := rand.Intn(maxAnz)
	fmt.Println(anz)
	for ; anz > 0; anz--{
		
		fmt.Println("for: anz Personen = äußere schleife:", anz)
			startEtage := rand.Intn(4)
			zielEtage := 0 
			fmt.Println("startetage: ",startEtage,"zieletage:",zielEtage)
			
			for ungleich := false; ungleich == false;{ // damit start und zieletage nicht gleich sind
				zielEtage = rand.Intn(4)
				fmt.Println("zieletage", zielEtage)
				if startEtage != zielEtage{
					ungleich = true
				}
			}
			neueAnfrage := Person{start: startEtage, ziel: zielEtage} // wie macht man einen channel ????????
			fmt.Println("Generiere Person",neueAnfrage)
			channelGP <- neueAnfrage
			//wg.Done()		
			
	}
	//wg.Done()

}






//---------------------------------------GOROUTINE PERSON-----------------------------------------------

func goroutineP(channelAlgPerson chan Person, channelAlgAufzug chan Aufzug,event string){

	aufzug := <- channelAlgAufzug
	fahrgast := <- channelAlgPerson
	
	switch event {
	case "aufzug bereit": 
		fahrgast.status = 1 // fahrgast hat jetzt einen zielaufzug
		fahrgast.aufzugNr = aufzug.aufzugNr
		fahrgast.soll = fahrgast.ziel - fahrgast.start
		
		if fahrgast.soll < 0{
			fahrgast.soll = fahrgast.soll *-1 // negative werte umkehren
		}
	case "aussteigen": 
		fahrgast.status = 3 // fahrgast als ausgestiegen markieren
		fahrgast.abweichung = fahrgast.ist - fahrgast.soll
	
	case "einsteigen":
		fahrgast.status = 2
		
	case "ziel unten":
		
		fahrgast.ist += 1
	case "ziel oben":
		
		fahrgast.ist += 1
	case "wartezeit":
		fahrgast.wartezeit += 1

	}
	channelAlgAufzug <- aufzug
	channelAlgPerson <- fahrgast 
	//wg.Done()
	//wg2.Done()

}

//------------------------------------------------------GOROUTINE AUFZUG------------------------------------------------------

func goroutineA(channelAlgPerson chan Person, channelAlgAufzug chan Aufzug,event string){

	aufzug := <- channelAlgAufzug
	fahrgast := <- channelAlgPerson
	
	switch event {
	case "aufzug bereit": 
		aufzug.zielEtage = fahrgast.start // aufzug erhält neues ziel

	case "aussteigen": 
		aufzug.fahrgaeste -= 1// anzahl der fahrgäste im Aufzug reduzieren

	case "einsteigen":
		aufzug.fahrgaeste += 1
	
	case "anfrage unten": 
		aufzug.etage -= 1
		aufzug.strecke += 1 // gesamtstrecke des aufzugs erhöhen
	
	case "anfrage oben":
		aufzug.etage += 1
		aufzug.strecke += 1 // gesamtstrecke des aufzugs erhöhen
	
	case "ziel unten":
		aufzug.etage -= 1
		aufzug.strecke += 1

	case "ziel oben":
		aufzug.etage += 1
		aufzug.strecke += 1


	}
	//wg.Done()
	//wg2.Done()

}


//----------------------------------------------ALGORITHMUS------------------------------------------------------------------

func Aufzugsteuerungs_Agorithmus_1 (channelAlgA chan []Aufzug, channelAlgP chan []Person, maxPers, dauer int){
	// bekommt anfragen von personen eine je person?
	// steuert wege der Aufzüge
	
	fmt.Println("ALGORITHMUS EINSTIEG")
	aufzugListe := make([]Aufzug,0)
	fahrgaesteListe := make([]Person,0)
	select{
	case aufzugListe = <- channelAlgA: aufzugListe = <- channelAlgA
	case fahrgaesteListe = <- channelAlgP: fahrgaesteListe = <- channelAlgP
	default:
	}
	
	fahrgast := make(chan Person)
	aufzug := make(chan Aufzug)

	for i := range aufzugListe {
		
		// planung 
		if aufzugListe[i].fahrgaeste == 0{
			//aufzug leer? dann erhalte neue anfrage
			for k:= range fahrgaesteListe{

				if fahrgaesteListe[k].status == 0{ // nehme ersten wartenden fahrgast
					//wg.Add(2)
					//wg2.Add(2)
					go goroutineP(fahrgast,aufzug,"aufzug bereit")
					go goroutineA(fahrgast,aufzug,"aufzug bereit")
					
					//wg.Wait()
					break
				} 

			}

		}

		// ausführung
		for j:= range fahrgaesteListe{


			// schritt: aussteigen lassen
			if aufzugListe[i].etage == fahrgaesteListe[j].ziel && fahrgaesteListe[j].status == 2 && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr{
				//wg.Add(2)
				//wg2.Add(2)
				go goroutineP(fahrgast,aufzug,"aussteigen")
				go goroutineA(fahrgast,aufzug,"aussteigen")
				fmt.Println("AAAAAAAAAAAAAAAAA")
				//wg.Wait()
				break

			// schritt: einsteigen lassen
			}else if aufzugListe[i].etage == fahrgaesteListe[j].start && fahrgaesteListe[j].status == 1 && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr{
				//wg.Add(2)
				//wg2.Add(2)
				go goroutineP(fahrgast,aufzug,"einsteigen")
				go goroutineA(fahrgast,aufzug,"einsteigen")
				//wg.wait()	
				break

				//schritt: zu anfrage runter fahren
			}else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug runter fahren
				//wg.Add(1)
				//wg2.Add(1)
				go goroutineA(fahrgast,aufzug,"anfrage unten")
			
				break

				//schritt: zu anfrage rauf fahren
			}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug rauf fahren
				//wg.Add(1)
				//wg2.Add(1)
				go goroutineA(fahrgast,aufzug,"anfrage oben")
				
				break
				//schritt: zu ziel des gasts runter fahren
			}else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{
				//wg.Add(2)
				//wg2.Add(2)
				go goroutineP(fahrgast,aufzug,"ziel unten")
				go goroutineA(fahrgast,aufzug,"ziel unten")
				
				break
				//schritt: zu ziel des gasts rauf fahren
			}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{
				//wg.Add(2)
				//wg2.Add(2)
				go goroutineP(fahrgast,aufzug,"ziel oben")
				go goroutineA(fahrgast,aufzug,"ziel oben")
				
				break
			}

			
			}
		
		}
		// wartezeiten der fahrgäste setzen
		for j := range fahrgaesteListe{
			if fahrgaesteListe[j].status == 0{ // wenn wartend, dann erhöhe wartezeit um 1
				//wg.Add(1)
				//wg2.Add(1)
				go goroutineP(fahrgast,aufzug,"wartezeit")
			
			

		}

		if dauer == 0{
			wg2.Done()
		}else{
			
			for j := range fahrgaesteListe{
				if fahrgaesteListe[j].status != 3{ // wenn noch nicht alle angekommen sind, abbrechen
				
					break
				}
				if j == len(fahrgaesteListe) - 1{
					wg2.Done()			// wenn die schleife bis hier hin gelaufen ist, sind alle angekommen
					break
				}
			
			

			}
		}

		

		channelAlgP <- fahrgaesteListe
		channelAlgA <- aufzugListe
		
	}
	//
	//wg2.Wait()
	//wg.Done()
}


//--------------------------------------------------------------MAIN---------------------------------------------
func main(){	
	go ZentraleSteuerlogik()
	wg.Add(1) // wartet auf ein done --> mehr dones = crash
	
	wg.Wait()
	

	
}

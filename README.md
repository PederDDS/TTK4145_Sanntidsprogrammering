# TTK4146 - Sanntidsprogrammering

Vi har implementert et peer to peer nettverk for et sett med heiser hvor alle tar avgjørelser basert på samme informasjon og samme grunnlag. Kommunikasjonen mellom heisene baserer seg på UDP-forbindelser, og meldingene som sendes er kart over heisene hvor all nødvendig informasjon for hver heis ligger lagret. 

## Modulinformasjon

### Nettverk
Har to funksjoner i det store og det hele. Det ene er å hele tiden oppdatere de andre heisene med endringene i det lokale kartet som ligger i ordermanager og oppdatere hverandre om at de lever ved å sende ut sin egen lokale ID over nettet. All kommunikasjon skjer med UDP. 

Handling av ordre: Dersom en heis får en ordre, vil den vente til det er registrert hos alle de andre heisene at det er en ordre i den etasjen, og den vil bli foreløpig akseptert. Dersom dette skjer for alle heisene, vil ordren settes til ORDER_ACCEPTED hos den heisen som blir tildelt ordren.

Oppdatering av heiser på nettverket: Hver enkelt heis sender ut sin egen ID i form av en string ved jevne mellomrom. Dette plukkes opp av de andre heisene. Dersom en heis faller ut av nettverket, vil det være umulig for den døde heisen å ta ordre, og de ordrene den hadde vil fordeles blant de resterende heisene. 



### FSM
Utfører logiske handlinger basert på hendelser i systemet, enten det er at en knapp trykkes, heisen ankommer en ny etasje, en timer er ferdig, eller at man mottar beskjed om at en heis på nettverket er død. Hver gang handlingene for en hendelse er utført, oppdateres det lokale kartet, som så sendes ut på nettverket. Kan både lese fra og skrive til ordermanager, i tillegg til å skrive til IO. 


### Ordermanager
Inneholder det lokale kartet over alle heisene. Informasjonen som ligger lagret for hver heis er tilstand, retning, etasje, ID og bestillinger. For hvert kart som kommer over nettverket eller hver gang FSM utfører handlinger som endrer på den lokale informasjonen til en heis, oppdateres det lokale kartet i ordermanager med den nye informasjonen. Det er også her bestillinger fordeles basert på en kostnadsfunksjon som er felles for alle heiser. På den måten vil alle heisene være enige om hvilken av dem som burde ta bestillinger.


### IO
Tar seg av interaksjon med den fysiske interfacen. Inneholder funksjoner for å sjekke om knapper er trykket og for å sette lys. Dersom en knapp blir trykket eller heisen ankommer en ny etasje, sendes dette over kanal til FSM. Inneholder også en del typedefinisjoner, eks. motorretning og type knapp. Denne informasjonen aksesseres av både main, FSM og ordermanager, men det er kun FSM som setter verdier på interfacen gjennom IO. 


### Def
Her ligger en del nyttige konstanter for heisen lagret, i tillegg til meldingsformatet for meldinger som sendes over kanalene. Informasjonen her kan aksesseres av alle de andre modulene.


## Bruk av annen kode
Systemet vårt bruker den utdelte nettverks- og driverkoden. Det er lagt til litt ekstra snacks i både peers og bcast i nettverksmodulen.
Heis-simulatoren som ligger vedlagt er også utdelt kode.

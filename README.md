# project-gruppa
project-gruppa created by GitHub Classroom

Vi har implementert et peer to peer nettverk for et sett med heiser hvor alle tar avgjørelser basert på samme informasjon og samme grunnlag. Kommunikasjonen mellom heisene baserer seg på UDP-forbindelser, og meldingene som sendes er kart over heisene hvor all nødvendig informasjon for hver heis ligger lagret. 

## Modulinformasjon

### Nettverk



### FSM
Utfører logiske handlinger basert på hendelser i systemet, enten det er at en knapp trykkes, heisen ankommer en ny etasje, en timer er ferdig, eller at man mottar beskjed om at en heis på nettverket er død. Hver gang handlingene for en hendelse er utført, oppdateres det lokale kartet, som så sendes ut på nettverket. 


### Ordermanager
Inneholder det lokale kartet over alle heisene. Informasjonen som ligger lagret for hver heis er tilstand, retning, etasje, ID og bestillinger. For hvert kart som kommer over nettverket eller hver gang FSM utfører handlinger som endrer på den lokale informasjonen til en heis, oppdateres det lokale kartet i ordermanager med den nye informasjonen. Det er også her bestillinger fordeles basert på en kostnadsfunksjon som er felles for alle heiser. På den måten vil alle heisene være enige om hvilken av dem som burde ta bestillinger.


### IO
Tar seg av interaksjon med den fysiske interfacen. Inneholder funksjoner for å sjekke om knapper er trykket og for å sette lys. Dersom en knapp blir trykket eller heisen ankommer en ny etasje, sendes dette over kanal til FSM.


### Def
Her ligger en del nyttige konstanter for heisen lagret.


## Nettverksdesign


### Feilhåndtering


## Kode som er lånt/fått
Systemet vårt bruker den utdelte nettverks- og driverkoden. Det er gjort noen endringer i nettverksmodulen fra det som ble utdelt.







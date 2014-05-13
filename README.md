To start the server
1. cd to "/src" and 
2. run command "go run main.go" for servers using paxos, "go run mainmencius.go" for servers uging mencius.

To start client
Open the browser and type in the IP address of the server followed by port 9999, i.e., 18.19.0.42:9999.

To delete all persistent storage for new canvas
Delete all storage files written. Both under src folder and projectserver folder

To run test
1. Go to respective folder, eg projectserver, menciusprojectserver  
2. run command "go test"
To run bench mark
Use command "go test -run=XXX -bench=."

Folder organization
src/resources contains js, css file for front-end application
src/projectserver contains servers using paxos in folder src/epaxos
src/menciusprojectserver contains servers using mencius in folder src/mencius

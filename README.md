# Little Alchemy Recipe Finder ⚗️
This project is deployed at https://tubes2festeicu-production.up.railway.app/. </br>
> This is a back-end repository. Please access front-end repository by clicking on [this link](https://github.com/wrdtlkhoir/Tubes2_FE_STEIcu.git).

## Overview
This project aims to get single or multiple recipe of elements in [Little Achemy 2](https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)). This project implements single and multiple recipes finder using Depth First Search (BFS), Breadth First Search (DFS), and Bidirectional Search. Multiple recipes finder is optimized using multithreading. The data used in the program is scrapped from the Little Alchemy 2 Website. User will input the method, algorithm, and number of recipes to be searched. The program will display the recipes found in form of solution tree including number of visited nodes, searching duration, and number of recipes found.

## Project Structure
```
├── 📁 doc
│   └── laporan.pdf
├── 📁 src
│   ├── bfs.go
│   ├── bidirection.go
│   ├── dfs.go
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── multiplebidirection.go
│   ├── package-lock.json
│   ├── recipes.json
│   ├── scraper.go
│   ├── test.html
│   ├── tree.go
│   └── treebidir.go
├── Dockerfile
├── README.md
├── docker-compose.yml
└── run.bat
```
## Algorithms
This section explains the algorithm used in this program, including BFS, DFS, and Bidirectional in a brief. Please refer to our [full report](./doc/) for more complete explaination and analysis.
### Breath First Search
For the single recipe method, the target element is the root node, and recipe combinations are enqueued for node checking. Each iteration dequeues the first element, checks if its node exists, and if not, creates and stores a new node. If the element is a basic element, the iteration moves to the next; otherwise, its recipe combinations are added to the queue. This process continues until the queue is empty, then the solution tree is constructed. For the multiple recipe method, BFS is performed using multithreading. Each recipe (with two ingredients) is processed in parallel with goroutines. These goroutines expand the recipe ingredients into tree nodes and explore sub-recipes recursively. The resulting nodes form paths to the target, collected via channels until the maximum path count is reached. Finally, the program converts the nodes into a tree structure and counts the elements in each path before returning the final result.

### Depth First Search
In the single recipe method, the dfsOne function recursively processes each element starting from the target's first combination. If the element is a basic ingredient, it's returned as a leaf node; otherwise, it recursively processes its components. A valid recipe tree is formed once a solution is found, and unique nodes visited during the search are recorded.
The multiple recipe DFS method optimizes the search using multithreading. Each recipe combination is explored in parallel with goroutines. The algorithm uses recursive DFS with cycle detection and depth limits. At shallow depths, DFS runs in parallel, while deeper levels are searched linearly. Only a subset of combinations is explored based on heuristics, and the results are combined into a unique solution tree.

### Bidirectional
The algorithm requires a tree structure representing all possible recipes from the target element to the base elements, built using a BFS approach. Two sets of data structures are initialized for bidirectional search: the Forward Search starts from the root node (target element) with a queue (q_f), a visited_f map for tracking visited nodes, and a forwardDepth map for node depth. The Backward Search starts simultaneously from all leaf nodes (base elements) with a second queue (q_b), a visited_b map for visited nodes, and a backwardDepth map for depth from the nearest base element. The search proceeds until both directions meet.

## Prerequisites
1. Go (version 1.24.2 or later)
   - Download and install Go from [go.dev](https://go.dev/dl/)
   - Verify the installation by running:
```
$ go version
```
3. Docker
   - Make sure you have Docker installed in your system. If you haven't, install [here](https://docs.docker.com/get-started/get-docker/)
  
## How to Compile and Run the Program
Clone this repository from terminal with this command:
```
$ git clone https://github.com/wrdtlkhoir/Tubes2_BE_STEIcu.git
```
### Run the application on development server
Compile the program by running the following command:
```
$ docker-compose up -d
```
### Run the application after doing updates
To run the program after doing updates, you can add a build tag by using this command
```
$ docker-compose up -d --build
```

## Available Scripts
In the project directory, you can run:
```
./run.bat
```
This runs the app in the development mode.

## Notes
This project uses the Go standard library packages:
- ```encoding/json``` for JSON encoding/decoding.
- ```fmt``` for formatting strings and printing to the console.
- ```log``` for logging errors.
- ```net/http``` for HTTP server handling.
- ```os``` for interacting with the file system.
- ```time``` for time-related operations.
- ```container/list``` for linked list data structures.
- ```sync``` for managing concurrency with goroutines and synchronization.
- ```context``` for managing request-scoped values, cancellation signals, and deadlines.

## Contributors 
<table>
  <tr>
    <td align="center">
      <a href="https://github.com/wrdtlkhoir">
        <img src="https://avatars.githubusercontent.com/wrdtlkhoir" width="80" style="border-radius: 50%;" /><br />
        <span><b>Wardatul Khoiroh </br> 13523001</b></span>
      </a>
    </td>
    <td align="center">
      <a href="https://github.com/najwakahanifatima">
        <img src="https://avatars.githubusercontent.com/najwakahanifatima" width="80" style="border-radius: 50%;" /><br />
        <span><b>Najwa Kahani Fatima </br> 13523043</b></span>
      </a>
    </td>
    <td align="center">
      <a href="https://github.com/numshv">
        <img src="https://avatars.githubusercontent.com/numshv" width="80" style="border-radius: 50%;" /><br />
        <span><b>Noumisyifa Nabila N. </br> 13523058</b></span>
      </a>
    </td>
  </tr>
</table>


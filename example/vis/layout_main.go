package main

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"regexp"
	"strings"
	"github.com/quan8/cofgra"
	dot "github.com/awalterschulze/gographviz"
)

var WIDTH = 15
var XGAP = 230
var YGAP = 80
var GENESIS = "v0x000000000000000000000000000000000000000000000000000000000GENESIS"

func readDotGraph(file string) (*dot.Graph) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil;
	    //Do something
	}

//	lines := strings.Split(string(content), "\n")
//	for _,s := range lines {
//		fmt.Println(s)
//	}

	ss := string(content)
	graphAst, _ := dot.ParseString(ss)
	graph := dot.NewGraph()
	if err := dot.Analyse(graphAst, graph); err != nil {
	    panic(err)
	}

	//output := graph.String()
	//fmt.Println(output)

	return graph
}

func addBottomNode(graph *dot.Graph) {
	graph.AddNode("GENESIS", GENESIS, nil)
	dotNodes := (*graph.Nodes).Nodes
	for i := 0; i < len(dotNodes); i++ {
		n := dotNodes[i]
		nName := (*n).Name

		if strings.HasSuffix(nName,"0000000000000000") {
			//fmt.Println("add genesis edge with", nName)
			graph.AddEdge(nName, GENESIS, true, nil)
		}
		if strings.HasSuffix(nName,"GENESIS") {
			(*n).Attrs["pos"] = "0,0"
			(*n).Attrs["label"] = "Genesis"
			(*n).Attrs["layer"] = "-1"
			(*n).Attrs["round"] = "-1"
		}
	}
}

func main() {
	// Read the file
	filename := "example/Node_127.0.0.1:12002.gv"
	dotgraph := readDotGraph(filename)
	//fmt.Println(dotgraph)

	dotNodeMap := make(map[string]*dot.Node)
	dotNodeMachineMap := make(map[string]int)

	graph := graff.NewEventGraph()

	r, _ := regexp.Compile("127.0.0.1:([0-9]+)")

	dotNodes := (*dotgraph.Nodes).Nodes
	for i := 0; i < len(dotNodes); i++ {
		n := dotNodes[i]
		nName := (*n).Name
		nAttrs := (*n).Attrs

		nLabel := nAttrs["label"]
		machineId := r.FindString(nLabel)
		//fmt.Println("node", nName, machineId)

		index, _ := strconv.Atoi(strings.Replace(machineId, "127.0.0.1:120", "", -1))
		//fmt.Println("node", nName, machineId, index)

		dotNodeMachineMap[nName] = index

		dotNodeMap[nName] = n
		graph.AddNode(nName)
	}

	dotEdges := (*dotgraph.Edges).Edges
	for i := 0; i < len(dotEdges); i++ {
		e := dotEdges[i]
		graph.AddEdge(e.Src, e.Dst)
	}

	//output := dotgraph.String()
	//fmt.Println(output)

	// Computes Coffman-Graham layering algorithm

//	dfs, err := graph.DFSSort()
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println("Depth-first search sort:", dfs)

	coffGra := graph.OptimizedCoffmanGrahamSorter(WIDTH)
	coffmanGraham, err := coffGra.EventSort();
	if err != nil {
		fmt.Println(err)
	}

//	fmt.Println("\n1st ===   Coffman-Graham sort (W=",WIDTH,"): ", coffmanGraham)
	for i := 0; i <len(coffmanGraham); i++  {
		layer := coffmanGraham[i]
		//fmt.Println("level=",i,", len=",len(layer)," ", layer);

		for j := 0; j <len(layer); j++  {
			nodeName := layer[j].(string)

			//fmt.Println("xxx", nodeName)

			node := dotNodeMap[nodeName]
			frame := node.Attrs["layer"]
			node.Attrs["layer"] = strconv.Itoa(i)
			node.Attrs["round"] = frame
			node.Attrs["cg"] = strconv.Itoa(i)

			x := dotNodeMachineMap[nodeName] * XGAP
			y := i * YGAP

			node.Attrs.Add("pos", fmt.Sprintf("\"%d,%d!\"", x, y));
		}
	}

	// Print out in dot format
	outputDot(coffmanGraham, dotNodeMap, dotEdges)
}

func outputDot(coffmanGraham [][]graff.Node, dotNodeMap map[string]*dot.Node, dotEdges []*dot.Edge) {
	fmt.Println("digraph {\n \tsplines=true;")
	for i := 0; i <len(coffmanGraham); i++  {
		layer := coffmanGraham[i]
		for j := 0; j <len(layer); j++  {
			nodeName := layer[j].(string)

			//fmt.Println("xxx", nodeName)

			node := dotNodeMap[nodeName]
			fmt.Printf("\t%s[ layer=%s, pos=%s, label=%s, round=%s, cg=%s, shape=none]\n", node.Name, node.Attrs["layer"],
				node.Attrs["pos"], node.Attrs["label"], node.Attrs["round"], node.Attrs["cg"])
		}
	}

	for i := 0; i < len(dotEdges); i++ {
		e := dotEdges[i]
		fmt.Printf("\t%s -> %s\n", e.Src, e.Dst);
	}

	fmt.Println("}")
}

func outputJson(coffmanGraham [][]graff.Node, dotNodeMap map[string]*dot.Node, dotEdges []*dot.Edge) {
	// print out in json format
//	{
//        name: 'graph2',
//        nodes: [
//            { id: 'node1', value: { label: 'node1' } },
//            { id: 'node2', value: { label: 'node2' } }
//        ],
//        links: [
//            { u: 'node1', v: 'node2', value: { label: 'link1' } }
//        ]
//	}

	fmt.Println("{")
	fmt.Println("\tname: 'graph3',")
	fmt.Println("\tnodes: [")
	for i := 0; i <len(coffmanGraham); i++  {
		layer := coffmanGraham[i]
		for j := 0; j <len(layer); j++  {
			nodeName := layer[j].(string)
			node := dotNodeMap[nodeName]
			fmt.Printf("\t\t{ id: '%s', value: {layer=%s, pos=%s, label=%s, round=%s, cg=%s}},\n", node.Name, node.Attrs["layer"],
				node.Attrs["pos"], node.Attrs["label"], node.Attrs["round"], node.Attrs["cg"])
		}
	}
	fmt.Println("\t],")
	fmt.Println("\tlinks: [")
	for i := 0; i < len(dotEdges); i++ {
		e := dotEdges[i]
		fmt.Printf("\t\t{u: '%s', v: '%s'},\n", e.Src, e.Dst);
	}
	fmt.Println("\t]")
	fmt.Println("}")
}

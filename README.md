#Backrunning bot

##Graph
Using graphtools (https://graph-tool.skewed.de/), I create a graph of tokens as nodes and LPs as edges by querying two subgraphs (via The Graph - https://thegraph.com/en/). These queries return the current reserves and token addresses needed to build the graph. Once the graph is built, I use a graphtools path finding algorithm to find all cycles of length 2 or 3. I then create a map of pair token address -> pair info (reserves and token address). I then save this map as a JSON file that will be used by my bot. Getting the latest graph should be run on startup of the bot. It only needs to be run on startup (not on-going).

##The bot
You need to provide a Blocknative API key, an RPC URL and the private key to your EOA
1. Loads the pairsToTokens.json file that we created above.
2. Uses the blocknative Golang SDK (https://github.com/marshabl/blocknative-go-sdk) to create a websocket feed to watch a number of different DEX and Aggregator addresses
3. Uses your RPC to watch for sync events on all token pairs in your graph, so that you get updated reserve data to be used in your profit calculations
4. Subscribes to block events to get the latest baseFee (although this could be done if Blocknative provides the baseFee in its payloads)
5. In mempool.go, I am leveraging Blocknative's sim platform and the net balance changes. For each transaction, I check net balance change addresses and if one of them is a LP in my graph, then I check for profit.
6. For each transaction I check for profit in utils.go. I am using the formula in Daniel's great post here (https://www.ddmckinnon.com/2022/11/27/all-is-fair-in-arb-and-mev-on-avalanche-c-chain/). So given a transaction that impacts a LP, I get all of the cycles from pairsToToken.json, and for each cycle I get the optimal amount in from the formula in the blog post. With this optimal amount in I check the profit of each cycle and I choose the highest one. If that is greater than baseFee * approx 150K gas, then I know this is a potential opportunity worth pursuing further
7. If there is a potential opportunity, I create the bundle and simulate it. If the simulation succeeds, I grab the actual gas used and calculate the actual gas price I want to spend based on a chosen margin. And then I fire it off to the builders through sendBundle
   
import json
from graph_tool.all import *
from subgraph import *
from web3 import Web3
from decimal import Decimal, getcontext

blacklistPairs = [
  '0xeBD49b4c8f7F0deD2cA8B951Cf92a583e7b4C8e7',
  '0x5a8bd8cdd4dA5aEbB783D2c67125bE0df484eEFe',
  '0x32a505BF9dB617d23bF3EbaAc9aeF80cB24a828C',
  '0x149148aCC3b06b8Cc73Af3A10E84189243A35925',
  '0xbE44E5dd67276F53b2Aca540B4eD32c84eD0e7B3',
  '0x2059024d050cdAfe4e4850596Ff1490cFc40c7Bd',
  '0xadeAa96A81eBBa4e3A5525A008Ee107385d588C3',
  '0xD3FF34FE9692c563bAfCb58c9A085C8C89c2F180',
  '0x17202e8E31FE0eeBdb401B2e11492F5Be52c8312',
  '0x6D7a7b251a6cAC43F1798494d847e6D333f9FFCa',
  '0x3150e4162DfDD5c89A8653bDf1CbB9E09C11F42a',
  '0x290CCAA8CC8e21d16f042d9AE0De9aC1805B824f',
  '0x20dd5Acc97E3f8685A9da6E499B21B22A7967279',
  '0xbaCd80B88104586D1CC9bcEa999781F3d393C332'

] #TODO

g = Graph(directed=False)
reserves0_prop = g.new_edge_property("string")
reserves1_prop = g.new_edge_property("string")
token0_prop = g.new_edge_property("string")
token1_prop = g.new_edge_property("string")
pair_prop = g.new_edge_property("string")
eprops = [reserves0_prop, reserves1_prop, token0_prop, token1_prop, pair_prop]
tokens = {}
vertices = {}
pairsToTokens = {}
tokensToPairs = {}
pairs = []

count = 0
for pair in uni:
  pairAddress = Web3.toChecksumAddress(pair['id'])
  if pairAddress not in blacklistPairs:
    reserves0 = pair['reserve0']
    decimals0 = pair['token0']['decimals']
    precision0 = len(reserves0)
    getcontext().prec = precision0
    reserves0 = hex(int(Decimal(reserves0)*(10**int(decimals0))))

    reserves1 = pair['reserve1']
    decimals1 = pair['token1']['decimals']
    precision1 = len(reserves1)
    getcontext().prec = precision1
    reserves1 = hex(int(Decimal(reserves1)*(10**int(decimals1))))
    token0 = Web3.toChecksumAddress(pair['token0']['id'])
    token1 = Web3.toChecksumAddress(pair['token1']['id'])

    if token0 not in tokens:
      tokens[token0] = count
      vertices[count] = token0
      g.add_vertex()
      count+=1
    if token1 not in tokens:
      tokens[token1] = count
      vertices[count] = token1
      g.add_vertex()
      count+=1

    v0 = g.vertex(tokens[token0])
    v1 = g.vertex(tokens[token1])
    g.add_edge_list([(v0, v1, reserves0, reserves1, token0, token1, pairAddress)], eprops=eprops)
    
    pairsToTokens[pairAddress] = {"pairInfo": {"token0": token0, "token1": token1, "reserves0": reserves0, "reserves1": reserves1}, "cycles": []}
    tokensToPairs["uni:"+token0+":"+token1] = pairAddress

for pair in sushi:
  pairAddress = Web3.toChecksumAddress(pair['id'])
  if pairAddress not in blacklistPairs:
    reserves0 = pair['reserve0']
    decimals0 = pair['token0']['decimals']
    precision0 = len(reserves0)
    getcontext().prec = precision0
    reserves0 = hex(int(Decimal(reserves0)*(10**int(decimals0))))

    reserves1 = pair['reserve1']
    decimals1 = pair['token1']['decimals']
    precision1 = len(reserves1)
    getcontext().prec = precision1
    reserves1 = hex(int(Decimal(reserves1)*(10**int(decimals1))))
    token0 = Web3.toChecksumAddress(pair['token0']['id'])
    token1 = Web3.toChecksumAddress(pair['token1']['id'])

    if token0 not in tokens:
      tokens[token0] = count
      vertices[count] = token0
      g.add_vertex()
      count+=1
    if token1 not in tokens:
      tokens[token1] = count
      vertices[count] = token1
      g.add_vertex()
      count+=1

    v0 = g.vertex(tokens[token0])
    v1 = g.vertex(tokens[token1])
    g.add_edge_list([(v0, v1, reserves0, reserves1, token0, token1, pairAddress)], eprops=eprops)
    
    pairsToTokens[pairAddress] = {"pairInfo": {"token0": token0, "token1": token1, "reserves0": reserves0, "reserves1": reserves1}, "cycles": []}
    tokensToPairs["sushi:"+token0+":"+token1] = pairAddress

paths = []
for path in all_paths(g, 1, 1, cutoff=4):
    if len(path) <= 4:
        if len(path) == 3:
            edges = g.edge(path[0], path[1], all_edges=True)
            if len(edges) == 2:
                paths.append(path)
        if len(path) == 4:
            paths.append(path)

for path in paths:
    if len(path) == 3:
        edges = g.edge(path[0], path[1], all_edges=True)
        if len(edges) == 2:
            pairAddress0 = pair_prop[edges[0]]
            pairAddress1 = pair_prop[edges[1]]
            
            cycle = [pairAddress0, pairAddress1]
            for pairAddress in cycle:
                if cycle not in pairsToTokens[pairAddress]['cycles']:
                    pairsToTokens[pairAddress]['cycles'].append(cycle)
            cycle = [pairAddress1, pairAddress0]
            for pairAddress in cycle:
                if cycle not in pairsToTokens[pairAddress]['cycles']:
                    pairsToTokens[pairAddress]['cycles'].append(cycle)
    else:
        cycle = []
        for i in range(len(path)-1):
            edge = g.edge(path[i], path[i+1])
            pairAddress = pair_prop[edge]
            cycle.append(pairAddress)
        for pairAddress in cycle:
            if cycle not in pairsToTokens[pairAddress]['cycles']:
                pairsToTokens[pairAddress]['cycles'].append(cycle)



for key, value in pairsToTokens.items():
  pairs.append(key)

with open('pairsToTokens.json', 'w') as f:
    json.dump(pairsToTokens, f)

with open('tokensToPairs.json', 'w') as f:
    json.dump(tokensToPairs, f)


import requests

sushi = []
uni = []
for i in range(3):
  skip = 1000*i
  variables = {"skip": skip}
  query = """
  query pairs($skip: Int!) {
    pairs(first: 1000, skip: $skip, orderBy: volumeUSD, orderDirection: desc, where: {
      txCount_gt: 10
    }) 
    {
      reserve0
      reserve1
      token0 {
        id
        decimals
      }
      token1 {
        id
        decimals
      }
      id
    }
  }
  """

  uniRes = requests.post(url="https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2", json={'query': query, 'variables': variables}).json()['data']['pairs']
  sushiRes = requests.post(url="https://api.thegraph.com/subgraphs/name/zippoxer/sushiswap-subgraph-fork", json={'query': query, 'variables': variables}).json()['data']['pairs']
  uni = uni + uniRes
  sushi = sushi + sushiRes
  pairsData = uni + sushi
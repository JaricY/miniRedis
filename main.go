package main

import "miniRedis/network"

var banner = `
   ______          ___
  / ____/___  ____/ (_)____
 / / __/ __ \/ __  / / ___/
/ /_/ / /_/ / /_/ / (__  )
\____/\____/\__,_/_/____/
`

func main() {
	print(banner)
	network.ElMain()
}

# tinyRouter

you  can run kubernetes cluster whit calico full bgp on L3  but not use  BGP Route Reflector and  not use Global BGP  peer 

auto add or del static route for pod subent by watching  calico's etcd

tinyRouter is default gateway for all kubernetes nodes


this a little different from the official plan

> https://projectcalico.docs.tigera.io/archive/v3.21/reference/architecture/design/l3-interconnect-fabric#bgp-only-interconnect-fabrics





tinyRouter will support hardware router later





## Physical Network traffic 


<img width="1016" alt="image" src="https://user-images.githubusercontent.com/47879545/174999103-a26d94e5-57e6-44e3-9898-7d0b324390b4.png">





-----------------



## Pod Network traffic 



<img width="1052" alt="image" src="https://user-images.githubusercontent.com/47879545/175005422-f7b293ab-b3ea-404f-a628-3327c2f0be4e.png">




# theoretical basis


## State machine for calico manage pod subnet 

....




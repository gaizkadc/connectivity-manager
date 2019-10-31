# connectivity-manager

This repository contains the code for the component connectivity-manager. The component resides in the Mngt Cluster and is in charge of managing the `status` of each App Cluster. It receives `ClusterAlive` checks from the different App Clusters and sets the clusters' status depending on these checks.

When a new platform is installed a `grace-period` is set (120 s by default) and the component takes 2 parameters: `threshold` and `offlinePolicy`.
* `threshold`: period after an App Cluster that has lost communication (stopped receiving `ClusterAlive` signals) with the Mngt Cluster will change its status from `ONLINE`/`ONLINE_CORDON`to `OFFLINE`/`OFFLINE_CORDON`.
* `offlinePolicy`: it defines the policy that will be triggered when a cluster has lost communication with the Mngt Cluster for a `grace-period` amount of time. It can be set to `none` or `drain`:
  * `none`: no policy will be triggered.
  * `drain`: the App Cluster will be drained (a `drain` signal will be sent to conductor) after the `grace-period` expires and all the applications running onn it will be redeployed somewhere else (when possible.)
  
##Cluster status lifecycle
* When an App Cluster is created and no `ClusterAlive signals` are being sent yet, the cluster status will be `UNKNOWN`.
* Once one of those checks arrives the connectivity-manager, its status will change to `ONLINE` for as long as the `ClusterAlive` signals are being received.
* If no check is received for longer than `threshold`, the cluster status will be set to `OFFLINE` if the previous status was `ONLINE` or `OFFLINE_CORDON` if the previous status was `ONLINE_CORDON`.
+ If the component doesn't get any `ClusterAlive` for longer than `grace-period`, the cluster status will be set to `OFFLINE_CORDON` and the `offlinePolicy` will be triggered.
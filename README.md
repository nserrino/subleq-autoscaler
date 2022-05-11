# subleq-autoscaler

A (toy) autoscaler for Kubernetes that can execute "subleq" instructions.  Since subleq is a [valid one-instruction set computer](https://en.wikipedia.org/wiki/One-instruction_set_computer), this means that any system capable of executing subleq commands is [Turing complete](https://en.wikipedia.org/wiki/Turing_completeness).

## Implementation

### Instruction execution
The [Kubernetes HorizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#how-does-a-horizontalpodautoscaler-work) runs its control loop by default every 15 seconds (though this demo app uses 30s as its period). Each invocation of the control loop will correspond to running a single `subleq` instruction in the input program.

### Program output
The "output" of the program is the number of pods in the deployment over time. In other words, the "tape" in the Turing machine will be the sequence of value for the number of pods present in the application.

| Invocation # | Time | Number of pods |
|--------------|------|----------------|
| 0            | 0s   | 1              |
| 1            | 15s  | 6              |
| 2            | 30s  | 2              |
| 3            | 45s  | 1              |

In this example, the output of the program could be read as `1,6,2,1`. However, we need to ensure that there is a distinct state for a terminated program. In order to ensure that the deployment has at least one pod, we will designate 1 pod as the "terminated" state for the program. In order to support an output value of 0 or greater, we will designate 2 as the # of pods corresponding to a 0 value in a running program. In other words, the output of the above program can actually be read as `-1,4,0,-1` where any value 0 and above corresponds to program output and `-1` indicates that the program either hasn't started running yet or has already terminated.

In this setup, deployments will always end with a single pod to indicate that the program has terminated.

### Input program
The input program to execute (consisting of `subleq` commands) will be inputed to the autoscaler in an encoding form, as the deployment name. A deployment's name can be decoded to an array of tuples. Each tuple contains 3 integers.

```
Pseudocode for a single instruction, `subleq a, b, c`. `r[i]` corresponds to the value at the `ith` register.
    r[b] = r[b] - r[a]
    if r[a] <= 0:
        jump to c
```

The "output" of the program, inspired by [this post](https://towardsdatascience.com/excel-fun-bd5a1a8992b8) comes from `r[a]` when `b` is `-1`, and `0` otherwise.

## Usage

### Prerequisites

You'll need to create a Kubernetes cluster that is capable of scheduling 107 new demo pods in order to run the demo.

### Trying it out

1. Deploy the autoscaler:

```
kubectl apply -f subleq-autoscaler-metrics.yaml
```

2. Deploy the demo app:

```
kubectl apply -f demo-app.yaml
```

3. Watch for the output, which will "print out" `Hi`:

```
# watch on Linux, or equivalent command on other OS.
watch -n 1 'kubectl get pods | grep Running | wc -l'
```

The number of running pods in the demo app should hit the following values:

| Time | Number of Pods | Corresponding Output Value | Notes |
|------|----------------|----------------------------|-------|
| 0s | 1 | -1 | `subleq` program hasn't started |
| 30s | 74 | 72 | Corresponding to ASCII `H` |
| 60s | 107 | 105 | Corresponding to ASCII `i` |
| 90s | 2 | 0 | Empty output |
| 120s+ | 1 | -1 | `subleq` program has terminated |

There will be a transition period between the output values as the number of pods rises and falls, so the output value should be sampled every 30s.

4. Try out your own demo app. You can provide your own subleq programs by modifying `demo-app.yaml` to contain instructions. See the code for how the values are parsed from the input strings.

### Cleaning up

1. Delete the `apiservice` resources corresponding to the autoscaler:
```
kubectl get apiservice --all-namespaces | grep subleq-autoscaler-metrics
kubectl delete apiservice <each value found in the above, e.g. v1beta1.custom.metrics.k8s.io>
```

2. Delete the `subleq-autoscaler-metrics` namespace:
```
kubectl delete namespace subleq-autoscaler-metrics
```

3. Delete the demo application:
```
kubectl delete -f demo-app.yaml
```

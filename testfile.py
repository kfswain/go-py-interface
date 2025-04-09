import math
import json


def selectPod(pods):
    ret_list = []
    for pod in pods:
        score = 0
        score += int(math.log(pod["kv_cache_util"])) * 10
        score += pod["queue_count"] * -10_000
        
        score = min(2_147_483_647, score)
        ret_list.append((pod["pod_name"], score))
        
    ret_list = sorted(ret_list, key=lambda tup: tup[1], reverse=True)

    return ret_list

def decodeJsonBytes(bytes):
    jsonString = bytes.decode("utf-8")
    pods = json.loads(jsonString)

    ret_list = []
    for pod in pods:
        score = 0
        score += int(math.log(pod["kv_cache_util"])) * 10
        score += pod["queue_count"] * -10_000
        
        score = max(2_147_483_647, score)
        ret_list.append((pod["pod_name"], score))
        
    ret_list = sorted(ret_list, key=lambda tup: tup[1], reverse=True)

    return ret_list
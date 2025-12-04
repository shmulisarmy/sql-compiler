person = {
	"name": "shmuli",
	"email": "email@gmail.com",
	"age": 25,
	"state": "state",
	"id": 1,
    "friends": ["berel", "shmuel", "shmuel"]
}


person1 = {
	"name": "berel",
	"email": "email@gmail.com",
	"age": "25",
	"state": "state",
	"id": 1,
    "friends": ["berel", "shmuel"]
}






def expect_array_and_actual_array_to_comparision_array(actual, expected):
    if len(actual) != len(expected):
        return {
            "expected-size": len(expected),
            "actual-size": len(actual),
        }
    comparison_result_tree = []
    for i in range(len(actual)):
        comparison_result_tree.append(comparison_result(actual[i], expected[i]))
    return comparison_result_tree



def comparison_result(expected, actual):
    if type(expected) != type(actual):
        return {
            "expected::type": type(expected).__name__,
            "actual::type": type(actual).__name__,
            "expected": expected,
            "actual": actual,
        }
    if type(expected) == list:
        return expect_array_and_actual_array_to_comparision_array(actual, expected)
    if type(expected) == dict:
        return expect_obj_and_actual_obj_to_comparision_obj(actual, expected)
    if expected == actual:
        return actual
    return {
        "expected": expected,
        "actual": actual,
    }


def expect_obj_and_actual_obj_to_comparision_obj(actual, expected):
    comparison_result_tree = {}
    for k, v in actual.items():
        if k in expected:
            comparison_result_tree[k] = comparison_result(v, expected[k])

    for k, v in expected.items():
        if k not in actual:
            comparison_result_tree[f"unexpected::{k}"] = v
    return comparison_result_tree

import json
print(json.dumps(expect_obj_and_actual_obj_to_comparision_obj(person1, person), indent=4))
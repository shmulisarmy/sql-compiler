import { createSignal, type Component } from 'solid-js';
import { live_db } from './live_db';
import { assert } from 'console';
import { deepObjectCompare } from './object_compare';






type Person={
  name:string
  email:string
  id:number
  todo:{[key: string]: {
      epic_title:string
      author:string
      id:number
    }}
}

const backend_base_url = "localhost:8080"

type Todo = Person['todo'][0]


const people: {[key: string]: Person} = live_db(`ws://${backend_base_url}/stream-data`);
const ws = new WebSocket(`ws://${backend_base_url}/stream-data`)




export function Todo_c({props}: {props:Todo}){
  return (
    <li class='flex flex-col space-y-2 ml-2 '>
      <p class="text-lg font-bold">{props.epic_title}</p>
      <p class="text-gray-500">{props.author}</p>
      <p class="text-xs text-gray-600">ID: {props.id}</p>
    </li>
  )
}
export function Person_c({props}: {props:Person}){
  return (
    <div class='flex flex-col space-y-2 p-1 m-1 border min-h-24'>
      <p class="text-lg font-bold">{props.name}</p>
      <p class="text-gray-500">{props.email}</p>
      <p class="text-xs text-gray-600">ID: {props.id}</p>
      <ul class='ml-2'>
        {Object.entries(props.todo).map(([id, todo]) => (
          <Todo_c  props={todo}/>
        ))}
      </ul>
    </div>
  )
}



async function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function run_tests(){
 //integration tests
    // The whole idea of this project is that the data is the same, regardless of the order
    // in which things were updated or data was received in. Therefore, we check that here
    // by having two different clients that are supposed to represent the same data. 
    // although the first source is created (and starts receiving data before updates are triggered, it should have be the same as the second one
  const client_data_reflection_1: {[key: string]: Person} = live_db(`ws://${backend_base_url}/stream-data`);
  fetch(`http://${backend_base_url}/add-person?name=donkey`).then(async  function(){
    const client_data_reflection_2: {[key: string]: Person} = live_db(`ws://${backend_base_url}/stream-data`);
    await sleep(200);
    const url = `https://json-diff-pro-copy-56b52272.base44.app/?actual=${encodeURIComponent(JSON.stringify(client_data_reflection_1))}&expected=${encodeURIComponent(JSON.stringify(client_data_reflection_2))}`
    // `https://compare-production-1494.up.railway.app/compare?expected=${encodeURIComponent(JSON.stringify(client_data_reflection_1))}&actual=${encodeURIComponent(JSON.stringify(client_data_reflection_2))}`
    console.log({url})
    if (!deepObjectCompare(client_data_reflection_1, client_data_reflection_2)){
      alert("test failed")
      console.log(`to see what is different between the 2 json objects follow ${url}`)
      throw new Error("test failed")
    } else {
      alert("test passed")
    }
  })
}

const App: Component = () => {
  return (
    <div>
      {JSON.stringify(people)}
      <button onClick={run_tests}>run_tests</button>
      <ul>
        {Object.entries(people).map(([id, person]) => (
          <Person_c  props={person}/>
        ))}
      </ul>
    </div>
  );
};

export default App;

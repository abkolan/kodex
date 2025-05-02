fn main() {
    let number: [i32;5] = [1,2,3,4,5];
    println!("Numbers Array = {:#?}",number);

    let fruits: [&str;3] = ["Apple","Banana","Orange"];
    println!("Fruits Array: {:?}", fruits);
    println!("Fruits Array: {:#?}", fruits);

    let human: (String,i32,bool) = ("Alice".to_string(), 30, false);
    println!("Human Tuple {:?}",human);

    let my_mix_tuple = ("Kratos", 23, true, [1,2,3]);
    println!("Mix Tuple {:?}", my_mix_tuple);
}   
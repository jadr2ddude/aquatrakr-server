int pir = 3;

void setup() {
  pinMode(pir, INPUT);
  Serial.begin(9600);

}

void loop() {
  int pirData = digitalRead(pir);
  Serial.println(pirData);

}

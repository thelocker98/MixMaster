#include <Adafruit_NeoPixel.h>

Adafruit_NeoPixel led(12, 2, NEO_GRB + NEO_KHZ800);




// --- Analog Sliders ---
const int NUM_SLIDERS = 5;
const int analogInputs[NUM_SLIDERS] = { A0, A1, A2, A3, A6 };
int analogSliderValues[NUM_SLIDERS];

// --- Buttons ---
const int NUM_BUTTONS = 10;
const int buttonPins[NUM_BUTTONS] = { 5, 6, 7, 8, 9, 10, 15, 16, 14, 3 };
int buttonStates[NUM_BUTTONS];



// the setup function runs once when you press reset or power the board
void setup() {
  // Initialize Serial
  Serial.begin(115200);

  // Initialize NeoPixel strip object (REQUIRED)
  led.begin();
  led.setBrightness(25);
}



// the loop function runs over and over again forever
void loop() {
  // Wait for a Second
  for (int i = 0; i < 12; i++)
    led.setPixelColor(i, led.Color(100, 200, 100));

  led.show();
  delay(100);

  for (int i = 0; i < 12; i++)
    led.setPixelColor(i, led.Color(0, 0, 0));
  led.show();
  delay(100);
  updateButtons();
  updateSliderValues();


  Serial.print("Slidders: ");
  for (int i = 0; i < NUM_SLIDERS; i++){
    Serial.print(analogSliderValues[i]);
    Serial.print(", ");
  }
  Serial.println();

  

  Serial.print("Buttons: ");
  for (int i = 0; i < NUM_BUTTONS; i++){
    Serial.print(buttonStates[i]);
    Serial.print(", ");
  }
  Serial.println();
}




void updateSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void updateButtons() {
  for (int i = 0; i < NUM_BUTTONS; i++) {
    bool button = digitalRead(buttonPins[i]);
    buttonStates[i] = button;
  }
}

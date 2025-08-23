// --- Sliders ---
const int NUM_SLIDERS = 5;
const int analogInputs[NUM_SLIDERS] = {A2, A0, A1, A5, A3};
int analogSliderValues[NUM_SLIDERS];

// --- Buttons ---
const int NUM_BUTTONS = 1;                  // change this if you add more
const int buttonPins[NUM_BUTTONS] = {10}; // add pins here
int buttonStates[NUM_BUTTONS];

void setup() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  for (int i = 0; i < NUM_BUTTONS; i++) {
    pinMode(buttonPins[i], INPUT_PULLUP);  // buttons wired to ground
  }

  Serial.begin(115200);
}

void loop() {
  updateSliderValues();
  updateButtons();
  sendJsonPacket();
  delay(10);
}

void updateSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void updateButtons() {
  for (int i = 0; i < NUM_BUTTONS; i++) {
    // invert because of INPUT_PULLUP (pressed = 0)
    buttonStates[i] = digitalRead(buttonPins[i]) == LOW ? 1 : 0;
  }
}

void sendJsonPacket() {
  Serial.print("{\"slidders\":[");
  for (int i = 0; i < NUM_SLIDERS; i++) {
    Serial.print(analogSliderValues[i] / 4); // scale 0–1023 → 0–255
    if (i < NUM_SLIDERS - 1) {
      Serial.print(",");
    }
  }
  Serial.print("], \"buttons\":[");
  for (int i = 0; i < NUM_BUTTONS; i++) {
    Serial.print(buttonStates[i]);
    if (i < NUM_BUTTONS - 1) {
      Serial.print(",");
    }
  }
  Serial.println("]}");
  delay(20);
}

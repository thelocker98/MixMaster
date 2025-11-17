String jsonData;

const String deviceSerial = "S0S04B00T0001";


// --- Sliders ---
const int NUM_SLIDERS = 5;
const int analogInputs[NUM_SLIDERS] = { A2, A0, A1, A5, A3 };
int analogSliderValues[NUM_SLIDERS];

// --- Buttons ---
const int NUM_BUTTONS = 1;                   // change this if you add more
const int buttonPins[NUM_BUTTONS] = { 10 };  // add pins here
int buttonStates[NUM_BUTTONS];
bool buttonTriggered[NUM_BUTTONS];



void setup() {
  // Setup Serial
  Serial.begin(115200);
  // Initiate Analog Sliders
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  // Initiate Buttions
  for (int i = 0; i < NUM_BUTTONS; i++) {
    pinMode(buttonPins[i], INPUT_PULLUP);
  }
}

void loop() {
  if (Serial.available() > 0) {       // Check if data is available
    char incoming = Serial.read();    // Read the incoming byte

    if (incoming == 'H') {            // If it's 'H', send data
      delay(20);
      sendData();
    }
  }
}


void sendData() {
  updateSliderValues();
  updateButtons();
  jsonData = createJsonPacket();
  Serial.println(jsonData);
}


void updateSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void updateButtons() {
  for (int i = 0; i < NUM_BUTTONS; i++) {
    bool button = digitalRead(buttonPins[i]);

    // invert because of INPUT_PULLUP (pressed = 0)
    if (!button && !buttonTriggered[i]) {
      buttonStates[i] = 1;
      buttonTriggered[i] = true;
    } else if (button) {
      buttonTriggered[i] = false;
      buttonStates[i] = 0;
    } else {
      buttonStates[i] = 0;
    }
  }
}

String createJsonPacket() {
  String data;
  data = "{\"id\":\"" + String(deviceSerial) + "\", ";
  data += "\"s\":[";
  for (int i = 0; i < NUM_SLIDERS; i++) {
    data += analogSliderValues[i] / 4;  // scale 0–1023 → 0–255
    if (i < NUM_SLIDERS - 1) {
      data += ",";
    }
  }
  data += "], \"b\":[";
  for (int i = 0; i < NUM_BUTTONS; i++) {
    data += buttonStates[i];
    if (i < NUM_BUTTONS - 1) {
      data += ",";
    }
  }
  data += "]}";
  return data;
}

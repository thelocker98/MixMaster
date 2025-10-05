String jsonData;

// Serial Number Generator
unsigned long makeSerial() {
  // Example: 2025-10-04 15:32:10 → 5100432
  const char* date = __DATE__;  // "Oct  4 2025"
  const char* time = __TIME__;  // "15:32:10"

  // Convert parts
  int month =
    (date[0] == 'J' && date[1] == 'a') ? 1 : (date[0] == 'F')                   ? 2
                                           : (date[0] == 'M' && date[2] == 'r') ? 3
                                           : (date[0] == 'A' && date[1] == 'p') ? 4
                                           : (date[0] == 'M' && date[2] == 'y') ? 5
                                           : (date[0] == 'J' && date[2] == 'n') ? 6
                                           : (date[0] == 'J' && date[2] == 'l') ? 7
                                           : (date[0] == 'A' && date[1] == 'u') ? 8
                                           : (date[0] == 'S')                   ? 9
                                           : (date[0] == 'O')                   ? 10
                                           : (date[0] == 'N')                   ? 11
                                                                                : 12;

  int day = ((date[4] == ' ') ? 0 : (date[4] - '0')) * 10 + (date[5] - '0');
  int hour = (time[0] - '0') * 10 + (time[1] - '0');
  int minute = (time[3] - '0') * 10 + (time[4] - '0');
  int second = (time[6] - '0') * 10 + (time[7] - '0');

  // Combine into a compact unique number (month*1000000 + day*10000 + hour*100 + minute)
  return (month * 1000000UL) + (day * 10000UL) + (hour * 100UL) + minute;
}

const unsigned long deviceSerial = makeSerial();






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
  delay(100);
  sendData();
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
  data = "{\"id\":" + String(deviceSerial) + ", ";
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
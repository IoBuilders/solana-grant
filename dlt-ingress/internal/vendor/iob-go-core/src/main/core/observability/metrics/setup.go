package metrics

const CommandMetricHistogram MetricName = "commands.duration"
const EventsMetricCounter MetricName = "events"
const FailuresMetricCounter MetricName = "failures"

const (
	FailureTypeAttr = "failureType"
)

func CoreMetricsSetup(registry *Registry) error {

	// Histogram for command
	histogram, err := Histogram(CommandMetricHistogram, "Duration of command execution in seconds", "s")

	if err != nil {
		return err
	}

	if err = registry.RegisterHistogram(CommandMetricHistogram, histogram); err != nil {
		return err
	}

	//Events counter
	counter, err := Counter(EventsMetricCounter, "Total events sent", "{event}")

	if err != nil {
		return err
	}
	if err = registry.RegisterCounter(EventsMetricCounter, counter); err != nil {
		return err
	}

	//failures counter
	failuresCounter, err := Counter(FailuresMetricCounter, "Total failures", "{failures}")

	if err != nil {
		return err
	}
	if err = registry.RegisterCounter(FailuresMetricCounter, failuresCounter); err != nil {
		return err
	}

	return nil
}

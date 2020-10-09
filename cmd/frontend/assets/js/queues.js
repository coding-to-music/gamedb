if ($('#queues-page').length > 0 || $('#player-missing-page').length > 0) {

    let activeWindow = true;

    $(window).on('focus', function () {
        activeWindow = true;
    });

    $(window).on('blur', function () {
        activeWindow = false;
    });

    const charts = {};
    $('[data-queue]').each(function (index, value) {
        charts[$(this).attr('data-queue')] = loadChart($(this).find('div').attr('id'));
    });

    updateCharts();

    const timer = window.setInterval(updateCharts, 10000); // 10 Seconds

    let firstRun = true;

    function updateCharts() {

        if (!activeWindow) {
            return;
        }

        $.ajax({
            url: '/queues/queues.json',
            dataType: 'json',
            cache: false,
            success: function (data, textStatus, jqXHR) {

                $.each(charts, function (index, value) {

                    let seriesKey = 0

                    for (let k in data) {
                        if (data.hasOwnProperty(k)) {
                            if (k.startsWith('GDB_' + index)) {

                                if (firstRun) {
                                    value.addSeries({
                                        name: k,
                                        data: data[k]['sum_messages'],
                                    });
                                } else {
                                    value.series[seriesKey].setData(data[k]['sum_messages']);
                                }

                                seriesKey++;
                            }
                        }
                    }
                });

                firstRun = false;
                $('#live-badge').addClass('badge-success').removeClass('badge-secondary badge-danger');
            },
            error: function (xhr, ajaxOptions, thrownError) {

                clearTimeout(timer);
                $('#live-badge').addClass('badge-danger').removeClass('badge-secondary badge-success');
                toast(false, 'Live functionality has stopped');
            }
        });
    }

    function loadChart(id) {

        return Highcharts.chart(id, $.extend(true, {}, defaultChartOptions, {
            chart: {
                animation: false,
            },
            legend: {
                enabled: false,
            },
            xAxis: {
                labels: {
                    step: 1,
                    formatter: function () {
                        return moment(this.value).format("h:mm");
                    },
                },
            },
            yAxis: {
                title: {
                    text: ''
                },
                allowDecimals: false,
                min: 0,
            },
            plotOptions: {
                series: {
                    marker: {
                        enabled: false // Too close together
                    },
                    animation: false
                }
            },
            series: [],
            tooltip: {
                formatter: function () {
                    return this.y.toLocaleString() + ' items in ' + this.series.name.replace(/^GDB_/, '') + ' at ' + moment(this.key).format("h:mm");
                },
            }
        }));
    }
}

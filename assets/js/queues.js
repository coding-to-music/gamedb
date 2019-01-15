if ($('#queues-page').length > 0) {

    let activeWindow = true;

    $(window).on('focus', function () {
        activeWindow = true;
    });

    $(window).on('blur', function () {
        activeWindow = false;
    });

    function updateChart() {

        if (!activeWindow) {
            return;
        }

        $.ajax({
            url: '/queues/ajax.json',
            success: function (data, textStatus, jqXHR) {
                chart.series[0].setData(data.items);
                chart.series[1].setData(data.rate);
            },
            dataType: 'json',
            cache: false,
            error: function (xhr, ajaxOptions, thrownError) {
                clearTimeout(timer);
                $('#live-badge').addClass('badge-danger').removeClass('badge-secondary badge-success');
                toast(false, 'Live functionality has stopped');
            }
        });
    }

    updateChart();
    const timer = window.setInterval(updateChart, 10000); // 10 Seconds

    const chart = Highcharts.chart('chart', {
        chart: {
            animation: false
        },
        title: {
            text: ''
        },
        subtitle: {
            text: ''
        },
        credits: {
            enabled: false
        },
        legend: {
            enabled: false
        },
        xAxis: {
            title: {
                text: 'Time'
            },
            labels: {
                step: 1,
                formatter: function () {
                    return moment(this.value).format("h:mm");
                },
            },
            type: 'datetime',
        },
        yAxis: [
            {
                title: {
                    text: 'Queue Size'
                },
                allowDecimals: false,
                min: 0,
            },
            {
                title: {
                    text: 'Queue Speed'
                },
                allowDecimals: false,
                min: 0,
                opposite: true,
            }
        ],
        plotOptions: {
            series: {
                marker: {
                    enabled: false // Too close together
                },
                animation: false
            }
        },
        series: [
            {
                color: '#28a745',
                yAxis: 0,
                name: 'size',
                type: 'areaspline',
            },
            {
                color: '#007bff',
                yAxis: 1,
                name: 'speed',
                type: 'spline'
            }
        ],
        tooltip: {
            formatter: function (x) {
                if (this.series.name === 'size') {
                    return this.y.toLocaleString() + ' items in the queue at ' + moment(this.key).format("h:mm");
                } else {
                    return this.y.toLocaleString() + ' items updated at ' + moment(this.key).format("h:mm");
                }
            },
        }
    });
}

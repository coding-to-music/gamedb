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
                    value.series[0].setData(data['GDB_' + index]['sum_messages']);
                });

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

        return Highcharts.chart(id, {
            chart: {
                animation: false,
                backgroundColor: 'rgba(0,0,0,0)',
            },
            title: {
                text: ''
            },
            subtitle: {
                text: ''
            },
            credits: {
                enabled: false,
            },
            legend: {
                enabled: false,
                itemStyle: {
                    color: '#28a745',
                },
                itemHiddenStyle: {
                    color: '#666666',
                },
            },
            xAxis: {
                title: {
                    text: ''
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
                        text: ''
                    },
                    allowDecimals: false,
                    min: 0,
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
            ],
            tooltip: {
                formatter: function () {
                    return this.y.toLocaleString() + ' items in the queue at ' + moment(this.key).format("h:mm") + ' UTC';
                },
            }
        });
    }
}
